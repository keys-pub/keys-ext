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
	key, err := s.parseKeyOrCurrent(req.KID)
	if err != nil {
		return nil, err
	}
	_, err = keys.NewUserForSigning(s.uc, key.ID(), req.Service, "test")
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
	key, err := s.parseKeyOrCurrent(req.KID)
	if err != nil {
		return nil, err
	}

	user, err := keys.NewUserForSigning(s.uc, key.ID(), req.Service, req.Name)
	if err != nil {
		return nil, err
	}
	msg, err := user.Sign(key.SignKey())
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
	key, err := s.parseKeyOrCurrent(req.KID)
	if err != nil {
		return nil, err
	}

	// Check if key is published (if not local user add)
	if !req.Local {
		resp, err := s.remote.Sigchain(key.ID())
		if err != nil {
			return nil, err
		}
		if resp == nil {
			return nil, errors.Errorf("key is not published")
		}
	}

	user, st, err := s.sigchainUserAdd(ctx, key, req.Service, req.Name, req.URL, req.Local)
	if err != nil {
		return nil, err
	}

	return &UserAddResponse{
		User:      userCheckToRPC(user),
		Statement: statementToRPC(st),
	}, nil
}

func (s *service) sigchainUserAdd(ctx context.Context, key keys.Key, service, name, url string, local bool) (*keys.UserCheck, *keys.Statement, error) {
	sc, err := s.scs.Sigchain(key.ID())
	if err != nil {
		return nil, nil, err
	}

	user, err := keys.NewUser(s.uc, key.ID(), service, name, url, sc.LastSeq()+1)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create user")
	}

	userCheck, err := s.uc.Check(ctx, user, key.PublicKey().SignPublicKey())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to check user")
	}

	st, err := keys.GenerateUserStatement(sc, user, key.SignKey(), s.Now())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to generate user statement")
	}

	if !local {
		// TODO: Check sigchain status (local changes not pushed)

		// Save to remote
		if s.remote == nil {
			return nil, nil, errors.Errorf("no remote set")
		}
		err := s.remote.PutSigchainStatement(st)
		if err != nil {
			return nil, nil, err
		}
	}
	if err := s.scs.AddStatement(st, key.SignKey()); err != nil {
		return nil, nil, err
	}
	return userCheck, st, nil
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

func userCheckToRPC(user *keys.UserCheck) *User {
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
		VerifiedAt: int64(keys.TimeToMillis(user.VerifiedAt)),
		Err:        user.Err,
	}
}

func userChecksToRPC(in []*keys.UserCheck) []*User {
	if in == nil {
		return nil
	}
	users := make([]*User, 0, len(in))
	for _, u := range in {
		users = append(users, userCheckToRPC(u))
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

func (s *service) findUser(ctx context.Context, kid keys.ID) (*keys.User, error) {
	sc, err := s.scs.Sigchain(kid)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		return nil, nil
	}
	users := sc.Users()
	if len(users) == 0 {
		return nil, nil
	}
	return users[len(users)-1], nil
}

func (s *service) searchUserByName(ctx context.Context, name string) (*keys.UserCheck, error) {
	if s.remote == nil {
		return nil, errors.Errorf("no remote set")
	}
	if strings.TrimSpace(name) != name {
		return nil, errors.Errorf("name has untrimmed whitespace")
	}
	if !strings.Contains(name, "@") {
		return nil, errors.Errorf("missing service")
	}
	resp, err := s.remote.Search(name, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(resp.Results) == 0 {
		return nil, nil
	}
	if len(resp.Results) > 1 {
		return nil, errors.Errorf("too many user results")
	}
	for _, user := range resp.Results[0].Users {
		if name == fmt.Sprintf("%s@%s", user.User.Name, user.User.Service) {
			return user, nil
		}
	}
	return nil, errors.Errorf("missing user in key")
}
