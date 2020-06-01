package service

import (
	"context"

	"github.com/keys-pub/keys/keyring"
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

func unlockPassword(kr *keyring.Keyring, password string) error {
	salt, err := kr.Salt()
	if err != nil {
		return err
	}
	key, err := keyring.KeyForPassword(password, salt)
	if err != nil {
		return err
	}
	if _, err := kr.Unlock(key); err != nil {
		return err
	}
	return nil
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
