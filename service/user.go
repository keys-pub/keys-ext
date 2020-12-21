package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/user"
	"github.com/keys-pub/keys/user/validate"
	"github.com/keys-pub/keys/users"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
			resp, err := s.client.User(ctx, kid)
			if err != nil {
				return nil, err
			}
			if resp != nil {
				res, err := s.updateUser(ctx, resp.User.KID, true)
				if err != nil {
					return nil, err
				}
				user = userResultToRPC(res)
			}
		}
	}

	return &UserResponse{
		User: user,
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
	key, err := s.edx25519Key(kid)
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
	key, err := s.edx25519Key(kid)
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
	key, err := s.edx25519Key(kid)
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

	linkService, err := validate.Lookup(service)
	if err != nil {
		return nil, nil, err
	}
	name = linkService.NormalizeName(name)
	urs, err = linkService.NormalizeURL(name, urs)
	if err != nil {
		return nil, nil, err
	}

	usr, err := user.New(key.ID(), service, name, urs, sc.LastSeq()+1)
	if err != nil {
		return nil, nil, err
	}

	userService, err := users.LookupService(usr, users.UseService(twitterProxy))
	if err != nil {
		return nil, nil, err
	}

	userResult := s.users.RequestVerify(ctx, userService, usr)
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
		err := s.client.SigchainSave(ctx, st)
		if err != nil {
			// TODO: Test this error response
			var rerr client.Error
			if errors.As(err, &rerr) && rerr.StatusCode == http.StatusConflict {
				return nil, nil, status.Error(codes.AlreadyExists, rerr.Message)
			}
			return nil, nil, err
		}
	}

	if err := s.scs.Save(sc); err != nil {
		return nil, nil, err
	}

	if _, err = s.users.Update(ctx, key.ID(), users.UseService(twitterProxy)); err != nil {
		return nil, nil, err
	}

	return userResult, st, nil
}

// SigchainURL is the sigchain URL for the user, or empty string if not set.
func (u *User) SigchainURL() string {
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

func userSearchResultsToRPC(results []*users.SearchResult) []*User {
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
		VerifiedAt: result.VerifiedAt,
		Timestamp:  result.Timestamp,
		Proxied:    result.Proxied,
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
	user := users[0]
	if user.ID != query {
		return nil, errors.Errorf("user search mismatch %s != %s", user.ID, query)
	}
	res, err := s.updateUser(ctx, keys.ID(user.KID), true)
	if err != nil {
		return nil, err
	}
	return userResultToRPC(res), nil
}

func (s *service) searchUsersLocal(ctx context.Context, query string, limit int) ([]*User, error) {
	query = strings.TrimSpace(query)
	logger.Infof("Search users local %q", query)
	res, err := s.users.Search(ctx, &users.SearchRequest{Query: query, Limit: limit})
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

func (s *service) ensureUserVerified(ctx context.Context, kid keys.ID) error {
	res, err := s.users.Get(ctx, kid)
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}

	if res.Status != user.StatusOK && res.Status != user.StatusConnFailure {
		return errors.Errorf("user %s has failed status %s", res.User.ID(), res.Status)
	}

	// Check if verified recently.
	if !res.IsVerifyExpired(s.clock.Now(), userVerifiedExpire) {
		return nil
	}

	// Our verify expired, re-check.
	logger.Infof("Checking user %v", res)
	resNew, err := s.updateUser(ctx, res.User.KID, true)
	if err != nil {
		return err
	}
	if resNew == nil {
		return errors.Errorf("failed user update: not found")
	}
	if resNew.Status != user.StatusOK {
		return errors.Errorf("user %s has failed status %s", resNew.User.ID(), resNew.Status)
	}
	return nil
}

// func (s *service) user(ctx context.Context, kid keys.ID, pull bool) (*User, error) {
// 	res, err := s.users.Get(ctx, kid)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if res == nil && pull {
// 		r, err := s.pullUser(ctx, kid)
// 		if err != nil {
// 			return nil, err
// 		}
// 		res = r
// 	}
// 	return userResultToRPC(res), nil
// }
