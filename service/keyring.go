package service

import (
	"github.com/keys-pub/keys-ext/git"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

func newKeyringStore(cfg *Config) (keyring.Store, error) {
	kt := cfg.Get(keyringTypeKey, "")
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

func repository(cfg *Config) (*git.Repository, error) {
	path, err := cfg.keyringGitPath()
	if err != nil {
		return nil, err
	}
	exists, err := pathExists(path)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	repo := git.NewRepository()
	// if err := repo.SetKey(key); err != nil {
	// 	return nil, err
	// }
	if err := repo.Open(path); err != nil {
		return nil, errors.Wrapf(err, "failed to open git repo")
	}
	return repo, nil
}
