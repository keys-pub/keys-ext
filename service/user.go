package service

import (
	"context"
	"fmt"
	strings "strings"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/user"
	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

// UserSearch (RPC) ...
func (s *service) UserSearch(ctx context.Context, req *UserSearchRequest) (*UserSearchResponse, error) {
	var users []*User
	if req.Local {
		u, err := s.searchUsersLocal(ctx, req.Query, int(req.Limit))
		if err != nil {
			return nil, err
		}
		users = u
	} else {
		u, err := s.searchUsersRemote(ctx, req.Query, int(req.Limit))
		if err != nil {
			return nil, err
		}
		users = apiUsersToRPC(u)
	}

	return &UserSearchResponse{
		Users: users,
	}, nil
}

// User (RPC) lookup user by kid.
func (s *service) User(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	if req.KID == "" {
		return nil, errors.Errorf("no kid specified")
	}
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}

	var user *User

	res, err := s.users.Get(ctx, kid)
	if err != nil {
		return nil, err
	}
	if res != nil {
		user = userResultToRPC(res)
	} else {
		if !req.Local {
			resp, err := s.remote.User(ctx, kid)
			if err != nil {
				return nil, err
			}
			if resp != nil {
				_, r, err := s.update(ctx, resp.User.KID)
				if err != nil {
					return nil, err
				}
				user = userResultToRPC(r)
			}
		}
	}

	return &UserResponse{
		User: user,
	}, nil
}

// UserService (RPC) validates a service.
func (s *service) UserService(ctx context.Context, req *UserServiceRequest) (*UserServiceResponse, error) {
	if req.Service == "" {
		return nil, errors.Errorf("no service specified")
	}
	key, err := s.parseSignKey(req.KID, true)
	if err != nil {
		return nil, err
	}
	_, err = user.NewUserForSigning(s.users, key.ID(), req.Service, "test")
	if err != nil {
		return nil, err
	}
	return &UserServiceResponse{Service: req.Service}, nil
}

// UserSign (RPC) creates a signed statement about a keys.
func (s *service) UserSign(ctx context.Context, req *UserSignRequest) (*UserSignResponse, error) {
	if req.Name == "" {
		return nil, errors.Errorf("no username specified")
	}
	if req.Service == "" {
		return nil, errors.Errorf("no service specified")
	}
	key, err := s.parseSignKey(req.KID, true)
	if err != nil {
		return nil, err
	}

	user, err := user.NewUserForSigning(s.users, key.ID(), req.Service, req.Name)
	if err != nil {
		return nil, err
	}
	msg, err := user.Sign(key)
	if err != nil {
		return nil, err
	}

	return &UserSignResponse{
		Message: msg,
		Name:    user.Name,
	}, nil
}

// UserAdd (RPC) adds a signed user statement to the sigchain.
func (s *service) UserAdd(ctx context.Context, req *UserAddRequest) (*UserAddResponse, error) {
	if req.Name == "" {
		return nil, errors.Errorf("no username specified")
	}
	if req.Service == "" {
		return nil, errors.Errorf("no service specified")
	}
	if req.URL == "" {
		return nil, errors.Errorf("no URL specified")
	}
	key, err := s.parseSignKey(req.KID, true)
	if err != nil {
		return nil, err
	}

	user, st, err := s.sigchainUserAdd(ctx, key, req.Service, req.Name, req.URL, req.Local)
	if err != nil {
		return nil, err
	}

	return &UserAddResponse{
		User:      userResultToRPC(user),
		Statement: statementToRPC(st),
	}, nil
}

func (s *service) sigchainUserAdd(ctx context.Context, key *keys.EdX25519Key, service, name, url string, localOnly bool) (*user.Result, *keys.Statement, error) {
	sc, err := s.scs.Sigchain(key.ID())
	if err != nil {
		return nil, nil, err
	}

	usr, err := user.NewUser(s.users, key.ID(), service, name, url, sc.LastSeq()+1)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create user")
	}

	userResult, err := s.users.Check(ctx, usr, key.ID())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to check user")
	}
	if userResult.Status != user.StatusOK {
		return nil, nil, errors.Errorf("failed to check user: %s", userResult.Err)
	}

	st, err := user.NewUserSigchainStatement(sc, usr, key, s.Now())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to generate user statement")
	}

	if err := sc.Add(st); err != nil {
		return nil, nil, err
	}

	if !localOnly {
		if err := s.remote.PutSigchainStatement(ctx, st); err != nil {
			return nil, nil, err
		}
	}

	if err := s.scs.SaveSigchain(sc); err != nil {
		return nil, nil, err
	}

	if _, err = s.users.Update(ctx, key.ID()); err != nil {
		return nil, nil, err
	}

	return userResult, st, nil
}

// SigchainURL is the sigchain URL for the user, or empty string if not set.
func (u User) SigchainURL() string {
	if u.Seq == 0 {
		return ""
	}
	return fmt.Sprintf("https://keys.pub/sigchain/%s/%d", u.KID, u.Seq)
}

func userStatus(s user.Status) UserStatus {
	switch s {
	case user.StatusUnknown:
		return UserStatusUnknown
	case user.StatusOK:
		return UserStatusOK
	case user.StatusResourceNotFound:
		return UserStatusResourceNotFound
	case user.StatusContentNotFound:
		return UserStatusContentNotFound
	case user.StatusConnFailure:
		return UserStatusConnFailure
	case user.StatusFailure:
		return UserStatusFailure
	default:
		panic(errors.Errorf("Unknown user status %s", s))
	}
}

func userSearchResultsToRPC(results []*user.SearchResult) []*User {
	users := make([]*User, 0, len(results))
	for _, r := range results {
		users = append(users, userResultToRPC(r.Result))
	}
	return users
}

func userResultToRPC(result *user.Result) *User {
	if result == nil {
		return nil
	}
	return &User{
		ID:         result.User.Name + "@" + result.User.Service,
		KID:        result.User.KID.String(),
		Seq:        int32(result.User.Seq),
		Service:    result.User.Service,
		Name:       result.User.Name,
		URL:        result.User.URL,
		Status:     userStatus(result.Status),
		VerifiedAt: int64(result.VerifiedAt),
		Timestamp:  int64(result.Timestamp),
		Err:        result.Err,
	}
}

func apiUsersToRPC(aus []*api.User) []*User {
	users := make([]*User, 0, len(aus))
	for _, au := range aus {
		users = append(users, apiUserToRPC(au))
	}
	return users
}

func apiUserToRPC(user *api.User) *User {
	if user == nil {
		return nil
	}
	return &User{
		ID:         user.ID,
		KID:        user.KID.String(),
		Seq:        int32(user.Seq),
		Service:    user.Service,
		Name:       user.Name,
		URL:        user.URL,
		Status:     userStatus(user.Status),
		VerifiedAt: int64(user.VerifiedAt),
		Timestamp:  int64(user.Timestamp),
		Err:        user.Err,
	}
}

func (s *service) searchRemoteCheckUser(ctx context.Context, userID string) (*User, error) {
	users, err := s.searchUsersRemote(ctx, userID, 1)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	user := users[0]
	if user.ID != userID {
		return nil, errors.Errorf("user search mismatch %s != %s", user.ID, userID)
	}
	_, r, err := s.update(ctx, keys.ID(user.KID))
	if err != nil {
		return nil, err
	}
	return userResultToRPC(r), nil
}

func (s *service) searchUsersLocal(ctx context.Context, query string, limit int) ([]*User, error) {
	query = strings.TrimSpace(query)
	logger.Infof("Search users local %q", query)
	res, err := s.users.Search(ctx, &user.SearchRequest{Query: query, Limit: limit})
	if err != nil {
		return nil, err
	}
	return userSearchResultsToRPC(res), nil
}

func (s *service) searchUsersRemote(ctx context.Context, query string, limit int) ([]*api.User, error) {
	query = strings.TrimSpace(query)
	logger.Infof("Search users remote %q", query)
	resp, err := s.remote.UserSearch(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}

func (s *service) parseIdentity(ctx context.Context, identity string, verify bool) (keys.ID, error) {
	return s.loadIdentity(ctx, identity, false, verify)
}

func (s *service) searchIdentity(ctx context.Context, identity string) (keys.ID, error) {
	return s.loadIdentity(ctx, identity, true, true)
}

func (s *service) loadIdentity(ctx context.Context, identity string, searchRemote bool, verify bool) (keys.ID, error) {
	if identity == "" {
		return "", errors.Errorf("no identity specified")
	}

	kid, err := s.findIdentity(ctx, identity, searchRemote)
	if err != nil {
		return "", err
	}

	if verify {
		if err := s.ensureVerified(ctx, kid); err != nil {
			return "", err
		}
	}

	return kid, nil
}

func (s *service) findIdentity(ctx context.Context, identity string, searchRemote bool) (keys.ID, error) {
	if identity == "" {
		return "", errors.Errorf("no identity specified")
	}

	if strings.Contains(identity, "@") {
		logger.Infof("Looking for user %q", identity)
		res, err := s.users.User(ctx, identity)
		if err != nil {
			return "", err
		}
		if res == nil {
			logger.Infof("User not found %s", identity)
			if searchRemote {
				user, err := s.searchRemoteCheckUser(ctx, identity)
				if err != nil {
					return "", err
				}
				if user == nil {
					return "", keys.NewErrNotFound(identity)
				}
				return keys.ID(user.KID), nil
			}
			return "", keys.NewErrNotFound(identity)
		}
		return res.User.KID, nil
	}

	id, err := keys.ParseID(identity)
	if err != nil {
		return "", errors.Errorf("failed to parse id %s", identity)
	}
	return id, nil
}

// userVerifiedExpire is how long a verify lasts.
// TOOD: Make configurable
const userVerifiedExpire = time.Hour * 24

// userCheckExpire is how long we wait between checks.
// TODO: Make configurable
const userCheckExpire = time.Hour * 24

// userCheckExpire is how long we wait between checks if not ok.
// TODO: Make configurable
const userCheckFailureExpire = time.Hour * 4

func (s *service) ensureVerified(ctx context.Context, kid keys.ID) error {
	res, err := s.users.Get(ctx, kid)
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}
	return s.ensureVerifiedResult(ctx, res)
}

func (s *service) ensureVerifiedResult(ctx context.Context, res *user.Result) error {
	if res == nil {
		return nil
	}

	if res.Status != user.StatusOK && res.Status != user.StatusConnFailure {
		return errors.Errorf("user %s has failed status %s", res.User.ID(), res.Status)
	}

	// Verified recently
	if !res.IsVerifyExpired(s.Now(), userVerifiedExpire) {
		return nil
	}

	// Our verify expired, re-check
	logger.Infof("Checking user %v", res)
	ok, resNew, err := s.update(ctx, res.User.KID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.Errorf("failed user update: not found")
	}
	if resNew.Status != user.StatusOK {
		return errors.Errorf("user %s has failed status %s", resNew.User.ID(), resNew.Status)
	}
	return nil
}

func (s *service) parseIdentities(ctx context.Context, recs []string, check bool) ([]keys.ID, error) {
	ids := make([]keys.ID, 0, len(recs))
	for _, r := range recs {
		id, err := s.parseIdentity(ctx, r, check)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *service) checkUpdateIfNeeded(ctx context.Context, kid keys.ID) error {
	res, err := s.users.Get(ctx, kid)
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}

	// If not OK, check every "userCheckFailureExpire", otherwise check every "userCheckExpire"
	now := s.Now()
	if (res.Status != user.StatusOK && res.IsTimestampExpired(now, userCheckFailureExpire)) ||
		res.IsTimestampExpired(now, userCheckExpire) {
		if _, _, err := s.update(ctx, kid); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) checkUpdate(ctx context.Context) error {
	logger.Infof("Checking keys...")
	pks, err := s.ks.EdX25519PublicKeys()
	if err != nil {
		return err
	}
	for _, pk := range pks {
		if err := s.checkUpdateIfNeeded(ctx, pk.ID()); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) updateAll(ctx context.Context) error {
	logger.Infof("Updating keys...")
	pks, err := s.ks.EdX25519PublicKeys()
	if err != nil {
		return err
	}
	for _, pk := range pks {
		if _, _, err := s.update(ctx, pk.ID()); err != nil {
			return err
		}
	}
	return nil
}
