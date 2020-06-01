package service

import (
	"context"
	"os"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// AuthSetup (RPC) ...
func (s *service) AuthSetup(ctx context.Context, req *AuthSetupRequest) (*AuthSetupResponse, error) {
	logger.Infof("Auth setup...")
	kr := s.keyring()
	status, err := kr.Status()
	if err != nil {
		return nil, err
	}
	if status != keyring.Setup {
		return nil, errors.Errorf("auth already setup")
	}

	if err := s.auth.setup(ctx, kr, req.Secret, req.Type); err != nil {
		return nil, err
	}

	// If setting up auth, and local database exists we should nuke it since the
	// pre-existing key is different. The database will be rebuilt on Open.
	path, err := s.cfg.AppPath(dbFilename, false)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Checking for existing db...")
	exists, err := pathExists(path)
	if err != nil {
		return nil, err
	}
	if exists {
		logger.Debugf("Removing existing db: %s", path)
		if err := os.RemoveAll(path); err != nil {
			return nil, err
		}
	}

	return &AuthSetupResponse{}, nil
}

// AuthUnlock (RPC) ...
func (s *service) AuthUnlock(ctx context.Context, req *AuthUnlockRequest) (*AuthUnlockResponse, error) {
	kr := s.keyring()
	status, err := kr.Status()
	if err != nil {
		return nil, err
	}
	if status == keyring.Setup {
		return nil, errors.Errorf("auth setup needed")
	}

	token, err := s.auth.unlock(ctx, kr, req.Secret, req.Type, req.Client)
	if err != nil {
		return nil, err
	}

	// TODO: Use a derived key instead of the actual key itself
	key := kr.MasterKey()
	if err := s.Open(ctx, key); err != nil {
		return nil, err
	}

	return &AuthUnlockResponse{
		AuthToken: token,
	}, nil
}

// AuthProvision (RPC) ...
func (s *service) AuthProvision(ctx context.Context, req *AuthProvisionRequest) (*AuthProvisionResponse, error) {
	kr := s.keyring()
	provision, err := s.auth.provision(ctx, kr, req.Secret, req.Type, req.Setup)
	if err != nil {
		return nil, err
	}
	return &AuthProvisionResponse{
		Provision: provisionToRPC(provision),
	}, nil
}

// AuthDeprovision (RPC) ...
func (s *service) AuthDeprovision(ctx context.Context, req *AuthDeprovisionRequest) (*AuthDeprovisionResponse, error) {
	kr := s.keyring()
	ok, err := kr.Deprovision(req.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, keys.NewErrNotFound(req.ID)
	}

	// TODO: If FIDO2 resident key and supports credMgmt remove from the device also?

	return &AuthDeprovisionResponse{}, nil
}

// AuthProvisions (RPC) ...
func (s *service) AuthProvisions(ctx context.Context, req *AuthProvisionsRequest) (*AuthProvisionsResponse, error) {
	kr := s.keyring()
	provisions, err := kr.Provisions()
	if err != nil {
		return nil, err
	}

	out := make([]*AuthProvision, 0, len(provisions))
	for _, provision := range provisions {
		out = append(out, provisionToRPC(provision))
	}

	return &AuthProvisionsResponse{
		Provisions: out,
	}, nil
}

// AuthLock (RPC) ...
func (s *service) AuthLock(ctx context.Context, req *AuthLockRequest) (*AuthLockResponse, error) {
	s.auth.reset()
	if err := s.keyring().Lock(); err != nil {
		return nil, err
	}

	s.Close()

	return &AuthLockResponse{}, nil
}

type testClientAuth struct {
	token string
}

func newTestClientAuth(token string) testClientAuth {
	return testClientAuth{token: token}
}

func (a testClientAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	if a.token == "" {
		return nil, nil
	}
	return map[string]string{
		"authorization": a.token,
	}, nil
}

func (a testClientAuth) RequireTransportSecurity() bool {
	// For test client
	return false
}

func provisionToRPC(p *keyring.Provision) *AuthProvision {
	return &AuthProvision{
		ID:        p.ID,
		Type:      authTypeToRPC(p.Type),
		AAGUID:    p.AAGUID,
		NoPin:     p.NoPin,
		CreatedAt: tsutil.Millis(p.CreatedAt),
	}
}

func authTypeToRPC(t keyring.AuthType) AuthType {
	switch t {
	case keyring.PasswordAuth:
		return PasswordAuth
	case keyring.FIDO2HMACSecretAuth:
		return FIDO2HMACSecretAuth
	default:
		return UnknownAuth
	}
}

func matchAAGUID(provisions []*keyring.Provision, aaguid string) *keyring.Provision {
	for _, provision := range provisions {
		if provision.AAGUID == aaguid {
			return provision
		}
	}
	return nil
}
