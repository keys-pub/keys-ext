package service

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

// KeyringFn provides a keyring.Keyring to the service.
type KeyringFn interface {
	// Keyring returns the service keyring.
	Keyring() *keyring.Keyring

	// Pull changes from remote (if supported).
	Pull() error

	// Push changes to remote (if supported).
	Push() error
}

func newKeyringFn(cfg *Config) (KeyringFn, error) {
	path, err := gitPath(cfg)
	if err != nil {
		return nil, err
	}
	if path != "" {
		return newGitKeyringFn(cfg)
	}
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

func newSystemKeyringFn(cfg *Config) (KeyringFn, error) {
	st, err := newKeyringStore(cfg)
	if err != nil {
		return nil, err
	}
	service := cfg.keyringService()
	sys, err := keyring.New(service, st)
	if err != nil {
		return nil, err
	}

	return &sysKeyringFn{sys: sys}, nil
}

func (k *sysKeyringFn) Keyring() *keyring.Keyring {
	return k.sys
}

func (k *sysKeyringFn) Pull() error {
	// Not supported by system keyring
	return nil
}

func (k *sysKeyringFn) Push() error {
	// Not supported by system keyring
	return nil
}

func newKeyringStore(cfg *Config) (keyring.Store, error) {
	kt := cfg.Get(keyringTypeCfgKey, "")
	switch kt {
	case "":
		logger.Infof("Keyring (default)")
		kr := keyring.SystemOrFS()
		logger.Infof("Keyring (default) using %s", kr.Name())
		return kr, nil
	case "fs":
		logger.Infof("Keyring (fs)")
		dir, err := cfg.AppPath("keyring", false)
		if err != nil {
			return nil, err
		}
		return keyring.FS(dir)
	case "mem":
		logger.Infof("Keyring (mem)")
		return keyring.Mem(), nil
	default:
		return nil, errors.Errorf("unknown keyring type %s", kt)
	}
}
