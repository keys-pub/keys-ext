package service

import (
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

func newKeyringStore(cfg *Config) (keyring.Store, error) {
	kt := cfg.Get(ckKeyringType, "")
	switch kt {
	case "":
		logger.Infof("Keyring (%s)", cfg.AppName())
		return keyring.System(), nil
	case "fs":
		logger.Infof("Keyring (store): fs")
		dir, err := cfg.AppPath("keyring", false)
		if err != nil {
			return nil, err
		}
		return keyring.NewFSStore(dir)
	case "mem":
		logger.Infof("Keyring (%s, store): mem", cfg.AppName())
		return keyring.NewMemStore(), nil
	default:
		return nil, errors.Errorf("unknown keyring type %s", kt)
	}
}
