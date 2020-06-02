package service

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

// keyringFn provides a keyring.Keyring to the service.
type keyringFn interface {
	// Keyring returns the service keyring.
	Keyring() *keyring.Keyring
}

func newKeyringFn(cfg *Config) (keyringFn, error) {
	return newSystemKeyringFn(cfg)
}

func (s *service) keyring() *keyring.Keyring {
	return s.keyringFn.Keyring()
}

func (s *service) keyStore() *keys.Store {
	return keys.NewStore(s.keyringFn.Keyring())
}

type sysKeyringFn struct {
	sys *keyring.Keyring
}

func newSystemKeyringFn(cfg *Config) (keyringFn, error) {
	st, err := newKeyringStore(cfg)
	if err != nil {
		return nil, err
	}
	sys, err := keyring.New(keyring.WithStore(st))
	if err != nil {
		return nil, err
	}

	return &sysKeyringFn{sys: sys}, nil
}

func (k *sysKeyringFn) Keyring() *keyring.Keyring {
	return k.sys
}

func newKeyringStore(cfg *Config) (keyring.Store, error) {
	kt := cfg.Get(keyringTypeCfgKey, "")
	switch kt {
	case "":
		logger.Infof("Keyring (default)")
		service := cfg.keyringService()
		st, err := keyring.NewSystemOrFS(service)
		if err != nil {
			return nil, err
		}
		logger.Infof("Keyring (default) using %s", st.Name())
		return st, nil
	case "fs":
		logger.Infof("Keyring (fs)")
		dir, err := cfg.AppPath("keyring", false)
		if err != nil {
			return nil, err
		}
		service := cfg.keyringService()
		return keyring.NewFS(service, dir)
	case "mem":
		logger.Infof("Keyring (mem)")
		return keyring.Mem(), nil
	default:
		return nil, errors.Errorf("unknown keyring type %s", kt)
	}
}
