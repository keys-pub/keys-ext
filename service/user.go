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
	_, err = keys.NewUserForSigning(key.ID(), req.Service, "test")
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

	usr, err := keys.NewUserForSigning(key.ID(), req.Service, req.Name)
	if err != nil {
		return nil, err
	}
	msg, err := usr.Sign(key.SignKey())
	if err != nil {
		return nil, err
	}

	return &UserSignResponse{
		Message: msg,
		Name:    usr.Name,
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

	usr, st, linkErr := s.sigchainUserAdd(ctx, key, req.Service, req.Name, req.URL, req.Local)
	if linkErr != nil {
		return nil, linkErr
	}

	return &UserAddResponse{
		User:      userToRPC(usr),
		Statement: statementToRPC(st),
	}, nil
}

func (s *service) sigchainUserAdd(ctx context.Context, key keys.Key, service, name, url string, local bool) (*keys.User, *keys.Statement, error) {
	sc, err := s.scs.Sigchain(key.ID())
	if err != nil {
		return nil, nil, err
	}

	usr, err := keys.NewUser(key.ID(), service, name, url, sc.LastSeq()+1)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create user")
	}

	if err := keys.UserCheckWithKey(ctx, usr, key.PublicKey().SignPublicKey(), keys.NewHTTPRequestor()); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to check user")
	}

	st, err := keys.GenerateUserStatement(sc, usr, key.SignKey(), s.Now())
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
	return usr, st, nil
}

// SigchainURL is the sigchain URL for the user, or empty string if not set.
func (u User) SigchainURL() string {
	if u.Seq == 0 {
		return ""
	}
	return fmt.Sprintf("https://keys.pub/sigchain/%s/%d", u.KID, u.Seq)
}

func userToRPC(usr *keys.User) *User {
	if usr == nil {
		return nil
	}
	return &User{
		KID:     usr.KID.String(),
		Seq:     int32(usr.Seq),
		Service: usr.Service,
		Name:    usr.Name,
		URL:     usr.URL,
	}
}

func usersToRPC(in []*keys.User) []*User {
	if in == nil {
		return nil
	}
	usrs := make([]*User, 0, len(in))
	for _, u := range in {
		usrs = append(usrs, userToRPC(u))
	}
	return usrs
}

func (s *service) findUser(ctx context.Context, kid keys.ID) (*keys.User, error) {
	sc, err := s.scs.Sigchain(kid)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		return nil, nil
	}
	usrs := sc.Users()
	if len(usrs) == 0 {
		return nil, nil
	}
	return usrs[len(usrs)-1], nil
}

func (s *service) findUserByName(ctx context.Context, name string) (*keys.User, error) {
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
	for _, usr := range resp.Results[0].Users {
		if name == fmt.Sprintf("%s@%s", usr.Name, usr.Service) {
			return usr, nil
		}
	}
	return nil, errors.Errorf("missing user in key")
}
