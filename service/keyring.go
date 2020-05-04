package service

import (
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

func newKeyringStore(cfg *Config) (keyring.Store, error) {
	kt := cfg.Get(keyringTypeKey, "")
	switch kt {
	case "":
		logger.Infof("Keyring (default)")
		kr := keyring.SystemOrFS()
		logger.Infof("Keyring, using %s", kr.Name())
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
		return keyring.NewMemStore(), nil
	default:
		return nil, errors.Errorf("unknown keyring type %s", kt)
	}
}
