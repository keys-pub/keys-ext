package service

import (
	"context"
	"fmt"
	strings "strings"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

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

func (s *service) sigchainUserAdd(ctx context.Context, key *keys.SignKey, service, name, url string, localOnly bool) (*keys.UserResult, *keys.Statement, error) {
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
		if err := s.remote.PutSigchainStatement(st); err != nil {
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

func userResultToRPC(result *keys.UserResult) *User {
	if result == nil {
		return nil
	}
	return &User{
		KID:        result.User.KID.String(),
		Seq:        int32(result.User.Seq),
		Service:    result.User.Service,
		Name:       result.User.Name,
		URL:        result.User.URL,
		Status:     userStatus(result.Status),
		VerifiedAt: int64(result.VerifiedAt),
		Err:        result.Err,
		Label:      result.User.Name + "@" + result.User.Service,
	}
}

func userToRPC(user *keys.User) *User {
	if user == nil {
		return nil
	}
	return &User{
		KID:     user.KID.String(),
		Seq:     int32(user.Seq),
		Service: user.Service,
		Name:    user.Name,
		URL:     user.URL,
		Status:  UserStatusUnknown,
	}
}

func usersToRPC(in []*keys.User) []*User {
	if in == nil {
		return nil
	}
	users := make([]*User, 0, len(in))
	for _, u := range in {
		users = append(users, userToRPC(u))
	}
	return users
}

func (s *service) searchUserExact(ctx context.Context, query string, local bool) (*keys.UserResult, error) {
	res, err := s.searchUser(ctx, query, 0, local)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res[0].UserResult, nil
}

func (s *service) searchUser(ctx context.Context, query string, limit int, local bool) ([]*keys.UserSearchResult, error) {
	if local {
		return s.searchUserLocal(ctx, query, limit)
	}
	return s.searchUserRemote(ctx, query, limit)
}

func (s *service) searchUserLocal(ctx context.Context, query string, limit int) ([]*keys.UserSearchResult, error) {
	query = strings.TrimSpace(query)
	return s.users.Search(ctx, &keys.UserSearchRequest{Query: query, Limit: limit})
}

func (s *service) searchUserRemote(ctx context.Context, query string, limit int) ([]*keys.UserSearchResult, error) {
	query = strings.TrimSpace(query)
	resp, err := s.remote.UserSearch(query, limit)
	if err != nil {
		return nil, err
	}
	return resp.Results, nil
}

func (s *service) parseIdentity(ctx context.Context, rec string) (keys.ID, error) {
	if rec == "" {
		return "", nil
	}
	if strings.Contains(rec, "@") {
		res, err := s.users.User(ctx, rec)
		if err != nil {
			return "", err
		}
		if res == nil {
			return "", keys.NewErrNotFound(rec)
		}
		if res.Status != keys.UserStatusOK {
			return "", errors.Errorf("user %s has failed status %s", rec, res.Status)
		}
		return res.User.KID, nil
	}

	id, err := keys.ParseID(rec)
	if err != nil {
		return "", errors.Errorf("failed to parse id  %s", rec)
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
