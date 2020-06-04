package service

import (
	"context"

	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

func setupPassword(kr *keyring.Keyring, password string) error {
	salt, err := kr.Salt()
	if err != nil {
		return err
	}
	key, err := keyring.KeyForPassword(password, salt)
	if err != nil {
		return err
	}
	provision := keyring.NewProvision(keyring.PasswordAuth)
	if err := kr.Setup(key, provision); err != nil {
		return err
	}
	return nil
}

func unlockPassword(kr *keyring.Keyring, password string) (*keyring.Provision, error) {
	salt, err := kr.Salt()
	if err != nil {
		return nil, err
	}
	key, err := keyring.KeyForPassword(password, salt)
	if err != nil {
		return nil, err
	}
	return kr.Unlock(key)
}

func provisionPassword(ctx context.Context, kr *keyring.Keyring, password string) (*keyring.Provision, error) {
	salt, err := kr.Salt()
	if err != nil {
		return nil, err
	}
	key, err := keyring.KeyForPassword(password, salt)
	if err != nil {
		return nil, err
	}
	provision := keyring.NewProvision(keyring.PasswordAuth)
	if err := kr.Provision(key, provision); err != nil {
		return nil, err
	}

	logger.Infof("Provision (password): %s", provision.ID)
	return provision, nil
}

// PasswordChange (RPC) ...
func (s *service) PasswordChange(ctx context.Context, req *PasswordChangeRequest) (*PasswordChangeResponse, error) {
	kr := s.keyring()
	old, err := unlockPassword(kr, req.Old)
	if err != nil {
		if errors.Cause(err) == keyring.ErrInvalidAuth {
			return nil, errors.Errorf("invalid password")
		}
		return nil, err
	}

	if _, err := provisionPassword(ctx, kr, req.New); err != nil {
		return nil, err
	}

	ok, err := kr.Deprovision(old.ID, false)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to deprovision old password (new password was provisioned)")
	}
	if !ok {
		return nil, errors.Errorf("failed to deprovision, password not found (new password was provisioned)")
	}
	return &PasswordChangeResponse{}, nil
}
