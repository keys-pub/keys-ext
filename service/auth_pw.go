package service

import (
	"context"
	"time"

	"github.com/keys-pub/keys/keyring"
)

func (a *auth) setupPassword(password string) error {
	salt, err := a.kr.Salt()
	if err != nil {
		return err
	}
	auth, err := keyring.NewPasswordAuth(password, salt)
	if err != nil {
		return err
	}
	if _, err := a.kr.Setup(auth); err != nil {
		return err
	}

	info := &authInfo{
		ID:        auth.ID(),
		Type:      passwordAuth,
		CreatedAt: time.Now(),
	}
	if err := a.saveInfo(info); err != nil {
		return err
	}

	return nil
}

func (a *auth) unlockPassword(password string) error {
	salt, err := a.kr.Salt()
	if err != nil {
		return err
	}
	auth, err := keyring.NewPasswordAuth(password, salt)
	if err != nil {
		return err
	}
	if _, err := a.kr.Unlock(auth); err != nil {
		return err
	}
	return nil
}

func (a *auth) provisionPassword(ctx context.Context, password string) (string, error) {
	salt, err := a.kr.Salt()
	if err != nil {
		return "", err
	}
	auth, err := keyring.NewPasswordAuth(password, salt)
	if err != nil {
		return "", err
	}
	id, err := a.kr.Provision(auth)
	if err != nil {
		return "", err
	}

	info := &authInfo{
		ID:        auth.ID(),
		Type:      passwordAuth,
		CreatedAt: time.Now(),
	}
	if err := a.saveInfo(info); err != nil {
		return "", err
	}

	logger.Infof("Provision (password) with auth id: %s", id)
	return string(id), nil
}
