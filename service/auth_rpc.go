package service

import (
	"context"
	"os"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

// AuthSetup (RPC) ...
func (s *service) AuthSetup(ctx context.Context, req *AuthSetupRequest) (*AuthSetupResponse, error) {
	logger.Infof("Auth setup...")
	status, err := s.ks.Keyring().Status()
	if err != nil {
		return nil, err
	}
	if status != keyring.Setup {
		return nil, errors.Errorf("auth already setup")
	}

	if err := s.auth.setup(ctx, req.Secret, req.Type); err != nil {
		return nil, err
	}

	// If setting up auth, and local database exists we should nuke it since the
	// pre-existing key is different. The database will be rebuilt on Open.
	path, err := s.cfg.AppPath(dbFilename, false)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Checking for existing db...")
	if _, err := os.Stat(path); err == nil {
		logger.Debugf("Removing existing db: %s", path)
		if err := os.RemoveAll(path); err != nil {
			return nil, err
		}
	}

	return &AuthSetupResponse{}, nil
}

// AuthUnlock (RPC) ...
func (s *service) AuthUnlock(ctx context.Context, req *AuthUnlockRequest) (*AuthUnlockResponse, error) {
	status, err := s.ks.Keyring().Status()
	if err != nil {
		return nil, err
	}
	if status == keyring.Setup {
		return nil, errors.Errorf("auth setup needed")
	}

	token, err := s.auth.unlock(ctx, req.Secret, req.Type, req.Client)
	if err != nil {
		return nil, err
	}

	// TODO: Use a derived key instead of the actual key itself
	key := s.auth.kr.MasterKey()
	if err := s.Open(ctx, key); err != nil {
		return nil, err
	}

	return &AuthUnlockResponse{
		AuthToken: token,
	}, nil
}

// AuthProvision (RPC) ...
func (s *service) AuthProvision(ctx context.Context, req *AuthProvisionRequest) (*AuthProvisionResponse, error) {
	id, err := s.auth.provision(ctx, req.Secret, req.Type, req.Setup)
	if err != nil {
		return nil, err
	}
	return &AuthProvisionResponse{
		ID: id,
	}, nil
}

// AuthDeprovision (RPC) ...
func (s *service) AuthDeprovision(ctx context.Context, req *AuthDeprovisionRequest) (*AuthDeprovisionResponse, error) {
	ok, err := s.auth.kr.Deprovision(req.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, keys.NewErrNotFound(req.ID)
	}

	// Remove auth info (ignore if it doesn't exist)
	if _, err := s.auth.deleteInfo(req.ID); err != nil {
		return nil, err
	}

	// TODO: If FIDO2 resident key and supports credMgmt remove from the device also?

	return &AuthDeprovisionResponse{}, nil
}

// AuthProvisions (RPC) ...
func (s *service) AuthProvisions(ctx context.Context, req *AuthProvisionsRequest) (*AuthProvisionsResponse, error) {
	ids, err := s.auth.kr.Provisions()
	if err != nil {
		return nil, err
	}

	provisions := make([]*AuthProvision, 0, len(ids))
	for _, id := range ids {
		if id == "v1.auth" {
			provisions = append(provisions, &AuthProvision{ID: id, Type: PasswordAuth})
			continue
		}

		info, err := s.auth.loadInfo(id)
		if err != nil {
			return nil, err
		}
		provision := &AuthProvision{ID: id}
		if info != nil {
			provision.Type = authTypeToRPC(info.Type)
			provision.AAGUID = info.AAGUID
			provision.NoPin = info.NoPin
		}
		provisions = append(provisions, provision)
	}

	return &AuthProvisionsResponse{
		Provisions: provisions,
	}, nil
}

// AuthLock (RPC) ...
func (s *service) AuthLock(ctx context.Context, req *AuthLockRequest) (*AuthLockResponse, error) {
	if err := s.auth.lock(); err != nil {
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
