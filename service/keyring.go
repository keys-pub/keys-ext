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

func checkKeyringConvert(env *Env, vlt *vault.Vault) error {
	empty, err := vlt.IsEmpty()
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	logger.Infof("Checking keyring conversion...")
	kr, err := newKeyring(env, "")
	if err != nil {
		return err
	}

	converted, err := vault.ConvertKeyring(kr, vlt)
	if err != nil {
		return errors.Wrapf(err, "failed to convert keyring")
	}
	if !converted {
		return nil
	}

	// Backup before resetting
	backupPath, err := env.AppPath(fmt.Sprintf("keyring-backup-%d.tgz", tsutil.Millis(time.Now())), false)
	if err != nil {
		return err
	}
	logger.Infof("Backing up keyring: %s", backupPath)
	if err := keyring.Backup(backupPath, kr, time.Now()); err != nil {
		return err
	}

	logger.Infof("Converted keyring, resetting...")
	if err := kr.Reset(); err != nil {
		return err
	}
	return nil
}

func newKeyring(env *Env, typ string) (keyring.Keyring, error) {
	switch typ {
	case "":
		// logger.Infof("Keyring (default)")
		st, err := defaultKeyring(env)
		if err != nil {
			return nil, err
		}
		logger.Debugf("Checking keyring (%s)", st.Name())
		return st, nil
	case "sys":
		service := keyringServiceName(env)
		return keyring.NewSystem(service)
	case "fs":
		logger.Debugf("Checking keyring (fs, deprecated)")
		dir, err := fsDir(env)
		if err != nil {
			return nil, err
		}
		st, err := keyring.NewFS(dir)
		if err != nil {
			return nil, err
		}
		return st, nil
	case "mem":
		logger.Debugf("Checking keyring (mem)")
		return keyring.NewMem(), nil
	default:
		return nil, errors.Errorf("unknown keyring type %s", typ)
	}
}

func defaultKeyring(env *Env) (keyring.Keyring, error) {
	// Check linux fallback.
	// We used to support a keyring type config option for "fs".
	// In earlier version of keyring, we used a fallback for linux at
	// ~/.keyring/<service>.
	if env.Get("keyring", "") == "fs" {
		return newKeyring(env, "fs")
	}
	if runtime.GOOS == "linux" {
		if err := keyring.CheckSystem(); err != nil {
			service := keyringServiceName(env)
			return linuxFallbackFS(service)
		}
	}

	// Use system
	return newKeyring(env, "sys")
}

func keyringServiceName(env *Env) string {
	return env.AppName() + ".keyring"
}

func linuxFallbackFS(service string) (keyring.Keyring, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(usr.HomeDir, ".keyring", service)
	return keyring.NewFS(dir)
}

func fsDir(env *Env) (string, error) {
	dir, err := env.AppPath("keyring", false)
	if err != nil {
		return "", err
	}
	service := keyringServiceName(env)
	return filepath.Join(dir, service), nil
}
