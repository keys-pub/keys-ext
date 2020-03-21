package service

import (
	"context"
	"fmt"
	strings "strings"

	"github.com/keys-pub/keys"
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
	_, err = keys.NewUserForSigning(s.users, key.ID(), req.Service, "test")
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

	user, err := keys.NewUserForSigning(s.users, key.ID(), req.Service, req.Name)
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

func (s *service) sigchainUserAdd(ctx context.Context, key *keys.EdX25519Key, service, name, url string, localOnly bool) (*keys.UserResult, *keys.Statement, error) {
	sc, err := s.scs.Sigchain(key.ID())
	if err != nil {
		return nil, nil, err
	}

	user, err := keys.NewUser(s.users, key.ID(), service, name, url, sc.LastSeq()+1)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create user")
	}

	userResult, err := s.users.Check(ctx, user, key.PublicKey())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to check user")
	}
	if userResult.Status != keys.UserStatusOK {
		return nil, nil, errors.Errorf("failed to check user: %s", userResult.Err)
	}

	st, err := keys.GenerateUserStatement(sc, user, key, s.Now())
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

func userStatus(s keys.UserStatus) UserStatus {
	switch s {
	case keys.UserStatusUnknown:
		return UserStatusUnknown
	case keys.UserStatusOK:
		return UserStatusOK
	case keys.UserStatusResourceNotFound:
		return UserStatusResourceNotFound
	case keys.UserStatusContentNotFound:
		return UserStatusContentNotFound
	case keys.UserStatusConnFailure:
		return UserStatusConnFailure
	case keys.UserStatusFailure:
		return UserStatusFailure
	default:
		panic(errors.Errorf("Unknown user status %s", s))
	}
}

func userSearchResultsToRPC(results []*keys.UserSearchResult) []*User {
	users := make([]*User, 0, len(results))
	for _, r := range results {
		users = append(users, userResultToRPC(r.UserResult))
	}
	return users
}

func userResultToRPC(result *keys.UserResult) *User {
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
	res, err := s.users.Search(ctx, &keys.UserSearchRequest{Query: query, Limit: limit})
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

func (s *service) parseIdentity(ctx context.Context, identity string) (keys.ID, error) {
	return s.loadIdentity(ctx, identity, false)
}

func (s *service) searchIdentity(ctx context.Context, identity string) (keys.ID, error) {
	return s.loadIdentity(ctx, identity, true)
}

func (s *service) loadIdentity(ctx context.Context, identity string, searchRemote bool) (keys.ID, error) {
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
		if res.Status != keys.UserStatusOK {
			return "", errors.Errorf("user %s has failed status %s", identity, res.Status)
		}
		return res.User.KID, nil
	}

	id, err := keys.ParseID(identity)
	if err != nil {
		return "", errors.Errorf("failed to parse id %s", identity)
	}
	return id, nil
}

func (s *service) parseIdentities(ctx context.Context, recs []string) ([]keys.ID, error) {
	ids := make([]keys.ID, 0, len(recs))
	for _, r := range recs {
		id, err := s.parseIdentity(ctx, r)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
