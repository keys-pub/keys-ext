package service

import (
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/syncp"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

func newKeyring(cfg *Config, typ string) (*keyring.Keyring, syncp.Config, error) {
	st, err := newKeyringStore(cfg, typ)
	if err != nil {
		return nil, syncp.Config{}, err
	}
	sys, err := keyring.New(keyring.WithStore(st))
	if err != nil {
		return nil, syncp.Config{}, err
	}

	return sys, syncp.Config{Dir: st.syncDir}, nil
}

type keyringStore struct {
	keyring.Store
	syncDir string
}

// When looking for the current keyring store, we check:
// - FS (versioned)
// - FS (deprecated)
// - Linux FS (fallback)
// - System
func defaultKeyringStore(cfg *Config) (*keyringStore, error) {
	// Check fs (versioned)
	fsv, err := hasFS(cfg, true)
	if err != nil {
		return nil, err
	}
	if fsv {
		return newKeyringStore(cfg, "fsv")
	}

	// Check fs
	fs, err := hasFS(cfg, false)
	if err != nil {
		return nil, err
	}
	if fs {
		return newKeyringStore(cfg, "fs")
	}

	// Check linux fallback.
	// In earlier version of keyring, we used a fallback for linux at
	// ~/.keyring/<service>.
	if runtime.GOOS == "linux" {
		if err := keyring.CheckSystem(); err != nil {
			service := keyringServiceName(cfg)
			return linuxFallbackFS(service)
		}
	}

	// Use system
	return newKeyringStore(cfg, "sys")
}

func newKeyringStore(cfg *Config, typ string) (*keyringStore, error) {
	switch typ {
	case "":
		logger.Infof("Keyring (default)")
		st, err := defaultKeyringStore(cfg)
		if err != nil {
			return nil, err
		}
		logger.Infof("Keyring (default) using %s", st.Name())
		return st, nil
	case "sys":
		service := keyringServiceName(cfg)
		st := keyring.NewSystem(service)
		return &keyringStore{st, ""}, nil
	case "fs":
		logger.Infof("Keyring (fs)")
		dir, err := fsDir(cfg, false)
		if err != nil {
			return nil, err
		}
		st, err := keyring.NewFS(dir, false)
		if err != nil {
			return nil, err
		}
		return &keyringStore{st, ""}, nil
	case "fsv":
		logger.Infof("Keyring (fsv)")
		dir, err := fsDir(cfg, true)
		if err != nil {
			return nil, err
		}
		st, err := keyring.NewFS(dir, true)
		if err != nil {
			return nil, err
		}
		return &keyringStore{st, dir}, nil
	case "mem":
		logger.Infof("Keyring (mem)")
		return &keyringStore{keyring.NewMem(), ""}, nil
	default:
		return nil, errors.Errorf("unknown keyring type %s", typ)
	}
}

type saltpackKeyring struct {
	*keyring.Keyring
}

func (k *saltpackKeyring) X25519Keys() ([]*keys.X25519Key, error) {
	return keys.X25519Keys(k.Keyring)
}

func migrateKeyring(cfg *Config, source string, destination string) error {
	from, err := newKeyringStore(cfg, source)
	if err != nil {
		return err
	}

	if destination == "" {
		return errors.Errorf("migrate destination is required")
	}

	to, err := newKeyringStore(cfg, destination)
	if err != nil {
		return err
	}

	if from.Name() == to.Name() {
		return errors.Errorf("migrate keyring source is same as destination %s == %s", from.Name(), to.Name())
	}

	// Migrate
	logger.Infof("Keyring copy from %s to %s ...", from.Name(), to.Name())
	ids, err := keyring.Copy(from, to)
	if err != nil {
		return err
	}
	logger.Infof("Keyring copied: %s", ids)

	// Backup and reset old keyring
	if _, err := backupKeyring(cfg, from); err != nil {
		return err
	}
	logger.Infof("Resetting old keyring...")
	if err := from.Reset(); err != nil {
		return err
	}

	return nil
}

func keyringServiceName(cfg *Config) string {
	return cfg.AppName() + ".keyring"
}

func hasFS(cfg *Config, versioned bool) (bool, error) {
	dir, err := fsDir(cfg, versioned)
	if err != nil {
		return false, err
	}
	return pathExists(dir)
}

func fsDir(cfg *Config, versioned bool) (string, error) {
	if versioned {
		dir, err := cfg.AppPath("keyring.fsv", false)
		if err != nil {
			return "", err
		}
		return dir, nil
	}
	dir, err := cfg.AppPath("keyring", false)
	if err != nil {
		return "", err
	}
	service := keyringServiceName(cfg)
	return filepath.Join(dir, service), nil
}

// linuxFallbackDir is a fallback used in an earlier version of the keyring.
func linuxFallbackDir(service string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".keyring", service), nil
}

func linuxFallbackFS(service string) (*keyringStore, error) {
	dir, err := linuxFallbackDir(service)
	if err != nil {
		return nil, err
	}
	st, err := keyring.NewFS(dir, false)
	if err != nil {
		return nil, err
	}
	return &keyringStore{st, ""}, nil
}
