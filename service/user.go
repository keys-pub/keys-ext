package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/link"
	"github.com/keys-pub/keys/user"
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

// Users (RPC) lookup users by kid.
func (s *service) Users(ctx context.Context, req *UsersRequest) (*UsersResponse, error) {
	if req.KID == "" {
		return nil, errors.Errorf("no kid specified")
	}
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}

	var users []*User

	usg, err := s.users.Get(ctx, kid)
	if err != nil {
		return nil, err
	}
	if usg != nil {
		users = userResultsToRPC(usg)
	} else {
		if !req.Local {
			resp, err := s.client.Users(ctx, kid)
			if err != nil {
				return nil, err
			}
			if resp != nil && len(resp.Users) > 0 {
				_, r, err := s.update(ctx, kid)
				if err != nil {
					return nil, err
				}
				users = userResultsToRPC(r)
			}
		}
	}

	return &UsersResponse{
		Users: users,
	}, nil
}

// UserService (RPC) validates a service.
func (s *service) UserService(ctx context.Context, req *UserServiceRequest) (*UserServiceResponse, error) {
	if req.KID == "" {
		return nil, errors.Errorf("no kid specified")
	}
	if req.Service == "" {
		return nil, errors.Errorf("no service specified")
	}
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}
	key, err := s.edX25519Key(kid)
	if err != nil {
		return nil, err
	}
	_, err = user.NewForSigning(key.ID(), req.Service, "test")
	if err != nil {
		return nil, err
	}
	return &UserServiceResponse{Service: req.Service}, nil
}

// UserSign (RPC) creates a signed statement about a keys.
func (s *service) UserSign(ctx context.Context, req *UserSignRequest) (*UserSignResponse, error) {
	if req.KID == "" {
		return nil, errors.Errorf("no kid specified")
	}
	if req.Name == "" {
		return nil, errors.Errorf("no username specified")
	}
	if req.Service == "" {
		return nil, errors.Errorf("no service specified")
	}
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}
	key, err := s.edX25519Key(kid)
	if err != nil {
		return nil, err
	}

	user, err := user.NewForSigning(key.ID(), req.Service, req.Name)
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
	if req.KID == "" {
		return nil, errors.Errorf("no kid specified")
	}
	if req.Name == "" {
		return nil, errors.Errorf("no name specified")
	}
	if req.Service == "" {
		return nil, errors.Errorf("no service specified")
	}
	if req.URL == "" {
		return nil, errors.Errorf("no URL specified")
	}
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}
	key, err := s.edX25519Key(kid)
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

func (s *service) sigchainUserAdd(ctx context.Context, key *keys.EdX25519Key, service, name, urs string, localOnly bool) (*user.Result, *keys.Statement, error) {
	sc, err := s.scs.Sigchain(key.ID())
	if err != nil {
		return nil, nil, err
	}

	linkService, err := link.NewService(service)
	if err != nil {
		return nil, nil, err
	}
	name = linkService.NormalizeName(name)
	urs, err = linkService.NormalizeURLString(name, urs)
	if err != nil {
		return nil, nil, err
	}

	usr, err := user.New(key.ID(), service, name, urs, sc.LastSeq()+1)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create user")
	}

	userResult := s.users.RequestVerify(ctx, usr)
	if userResult.Status != user.StatusOK {
		return nil, nil, errors.Errorf("user check failed: %s", userResult.Err)
	}

	st, err := user.NewSigchainStatement(sc, usr, key, s.clock.Now())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to generate user statement")
	}

	if err := sc.Add(st); err != nil {
		return nil, nil, err
	}

	if !localOnly {
		if err := s.client.SigchainSave(ctx, st); err != nil {
			return nil, nil, err
		}
	}

	if err := s.scs.Save(sc); err != nil {
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
	case user.StatusOK:
		return UserStatusOK
	case user.StatusResourceNotFound:
		return UserStatusResourceNotFound
	case user.StatusContentNotFound:
		return UserStatusContentNotFound
	case user.StatusConnFailure:
		return UserStatusConnFailure
	case user.StatusContentInvalid:
		return UserStatusContentInvalid
	case user.StatusFailure:
		return UserStatusFailure
	default:
		return UserStatusUnknown
	}
}

func userSearchResultsToRPC(results []*user.SearchResult) []*User {
	users := make([]*User, 0, len(results))
	for _, r := range results {
		users = append(users, userResultToRPC(r.Result))
	}
	return users
}

func userResultsToRPC(results []*user.Result) []*User {
	users := make([]*User, 0, len(results))
	for _, r := range results {
		users = append(users, userResultToRPC(r))
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
		VerifiedAt: result.VerifiedAt,
		Timestamp:  result.Timestamp,
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
		VerifiedAt: user.VerifiedAt,
		Timestamp:  user.Timestamp,
		Err:        user.Err,
	}
}

func (s *service) searchRemoteCheckUser(ctx context.Context, query string) (*User, error) {
	users, err := s.searchUsersRemote(ctx, query, 1)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	usr := users[0]
	if usr.ID != query {
		return nil, errors.Errorf("user search mismatch %s != %s", usr.ID, query)
	}
	_, res, err := s.update(ctx, usr.KID)
	if err != nil {
		return nil, err
	}
	result := findResultForAPIUser(usr, res)
	if result == nil {
		return nil, nil
	}
	return userResultToRPC(result), nil
}

func findResultForAPIUser(usr *api.User, results []*user.Result) *user.Result {
	for _, res := range results {
		if res.User.ID() == usr.ID {
			return res
		}
	}
	return nil
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
	resp, err := s.client.UserSearch(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}

func (s *service) lookupUser(ctx context.Context, user string, searchRemote bool) (keys.ID, error) {
	if searchRemote {
		result, err := s.searchRemoteCheckUser(ctx, user)
		if err != nil {
			return "", err
		}
		if result == nil {
			return "", keys.NewErrNotFound(user)
		}
		return keys.ID(result.KID), nil
	}
	res, err := s.users.User(ctx, user)
	if err != nil {
		return "", err
	}
	if res == nil {
		return "", keys.NewErrNotFound(user)
	}
	return res.User.KID, nil
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

func (s *service) ensureUsersVerified(ctx context.Context, kid keys.ID) error {
	res, err := s.users.Get(ctx, kid)
	if err != nil {
		return err
	}

	needsVerify := false
	for _, r := range res {
		if r.Status != user.StatusOK && r.Status != user.StatusConnFailure {
			return errors.Errorf("user %s has failed status %s", r.User.ID(), r.Status)
		}

		// Check if verified recently.
		if !r.IsVerifyExpired(s.clock.Now(), userVerifiedExpire) {
			continue
		}

		needsVerify = true
		break
	}

	if needsVerify {
		// Our verify expired, re-check.
		ok, rup, err := s.update(ctx, kid)
		if err != nil {
			return err
		}
		if !ok {
			return errors.Errorf("failed user update: not found")
		}
		for _, r := range rup {
			if r.Status != user.StatusOK {
				return errors.Errorf("user %s has failed status %s", r.User.ID(), r.Status)
			}
		}
	}
	return nil
}
