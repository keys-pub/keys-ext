package service

import (
	"fmt"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

func checkKeyringConvert(cfg *Config, vlt *vault.Vault) error {
	empty, err := vlt.IsEmpty()
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	logger.Infof("Checking keyring convert...")
	kr, err := newKeyring(cfg, "")
	if err != nil {
		return err
	}

	// Backup
	backupPath, err := cfg.AppPath(fmt.Sprintf("keyring-backup-%d.tgz", tsutil.Millis(time.Now())), false)
	if err != nil {
		return err
	}
	logger.Infof("Backing up keyring: %s", backupPath)
	if err := keyring.Backup(backupPath, kr, time.Now()); err != nil {
		return err
	}

	if err := vault.ConvertKeyring(kr, vlt); err != nil {
		return errors.Wrapf(err, "failed to convert keyring")
	}
	logger.Infof("Converted keyring, resetting...")
	if err := kr.Reset(); err != nil {
		return err
	}
	return nil
}

func newKeyring(cfg *Config, typ string) (keyring.Keyring, error) {
	switch typ {
	case "":
		logger.Infof("Keyring (default)")
		st, err := defaultKeyring(cfg)
		if err != nil {
			return nil, err
		}
		logger.Infof("Keyring using %s", st.Name())
		return st, nil
	case "sys":
		service := keyringServiceName(cfg)
		return keyring.NewSystem(service)
	case "fs":
		logger.Infof("Keyring (fs, deprecated)")
		dir, err := fsDir(cfg)
		if err != nil {
			return nil, err
		}
		st, err := keyring.NewFS(dir)
		if err != nil {
			return nil, err
		}
		return st, nil
	case "mem":
		logger.Infof("Keyring (mem)")
		return keyring.NewMem(), nil
	default:
		return nil, errors.Errorf("unknown keyring type %s", typ)
	}
}

func defaultKeyring(cfg *Config) (keyring.Keyring, error) {
	// Check linux fallback.
	// We used to support a keyring type config option for "fs".
	// In earlier version of keyring, we used a fallback for linux at
	// ~/.keyring/<service>.
	if cfg.Get("keyring", "") == "fs" {
		return newKeyring(cfg, "fs")
	}
	if runtime.GOOS == "linux" {
		if err := keyring.CheckSystem(); err != nil {
			service := keyringServiceName(cfg)
			return linuxFallbackFS(service)
		}
	}

	// Use system
	return newKeyring(cfg, "sys")
}

func keyringServiceName(cfg *Config) string {
	return cfg.AppName() + ".keyring"
}

func linuxFallbackFS(service string) (keyring.Keyring, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(usr.HomeDir, ".keyring", service)
	return keyring.NewFS(dir)
}

func fsDir(cfg *Config) (string, error) {
	dir, err := cfg.AppPath("keyring", false)
	if err != nil {
		return "", err
	}
	service := keyringServiceName(cfg)
	return filepath.Join(dir, service), nil
}
