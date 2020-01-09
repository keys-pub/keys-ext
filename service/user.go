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
	key, err := s.parseKey(req.KID)
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
	key, err := s.parseKey(req.KID)
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
	key, err := s.parseKey(req.KID)
	if err != nil {
		return nil, err
	}

	user, st, err := s.sigchainUserAdd(ctx, key, req.Service, req.Name, req.URL, req.LocalOnly)
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

func userResultToRPC(user *keys.UserResult) *User {
	if user == nil {
		return nil
	}
	return &User{
		KID:        user.User.KID.String(),
		Seq:        int32(user.User.Seq),
		Service:    user.User.Service,
		Name:       user.User.Name,
		URL:        user.User.URL,
		Status:     userStatus(user.Status),
		VerifiedAt: int64(user.VerifiedAt),
		Err:        user.Err,
	}
}

func userResultsToRPC(in []*keys.UserResult) []*User {
	if in == nil {
		return nil
	}
	users := make([]*User, 0, len(in))
	for _, u := range in {
		users = append(users, userResultToRPC(u))
	}
	return users
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

func (s *service) searchUserLocalExact(ctx context.Context, query string) (*keys.UserResult, error) {
	res, err := s.searchUserLocal(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}

	for _, user := range res[0].Users {
		if query == fmt.Sprintf("%s@%s", user.User.Name, user.User.Service) {
			return user, nil
		}
	}

	return nil, nil
}

func (s *service) searchUserExact(ctx context.Context, query string) (*keys.UserResult, error) {
	res, err := s.searchUser(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}

	for _, user := range res[0].Users {
		if query == fmt.Sprintf("%s@%s", user.User.Name, user.User.Service) {
			return user, nil
		}
	}

	return nil, nil
}

func (s *service) searchUser(ctx context.Context, query string) ([]*keys.SearchResult, error) {
	res, err := s.searchUserLocal(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return res, nil
	}
	return s.searchUserRemote(ctx, query)
}

func (s *service) searchUserLocal(ctx context.Context, query string) ([]*keys.SearchResult, error) {
	query = strings.TrimSpace(query)
	return s.users.Search(ctx, &keys.SearchRequest{Query: query})
}

func (s *service) searchUserRemote(ctx context.Context, query string) ([]*keys.SearchResult, error) {
	query = strings.TrimSpace(query)
	resp, err := s.remote.Search(query, 0)
	if err != nil {
		return nil, err
	}
	return resp.Results, nil
}
