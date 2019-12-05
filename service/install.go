package service

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func defaultServicePath() string {
	exe, exeErr := ExecutablePath()
	if exeErr != nil {
		panic(exeErr)
	}
	dir := filepath.Dir(exe)

	name := "keysd"
	if runtime.GOOS == "windows" {
		name = "keysd.exe"
	}

	servicePath := filepath.Join(dir, name)
	return servicePath
}

func restart(cfg *Config) error {
	return restartProcess(cfg)
}

func start(cfg *Config, wait bool) error {
	if err := startProcess(cfg); err != nil {
		return err
	}
	if wait {
		if err := waitForStart(cfg); err != nil {
			return err
		}
	}
	return nil
}

func stop(cfg *Config) error {
	if err := stopProcess(cfg); err != nil {
		return err
	}
	pidPath, err := cfg.AppPath("pid", false)
	if err != nil {
		return err
	}
	if err := removeFile(pidPath); err != nil {
		return err
	}

	return nil
}

func removeFile(pidPath string) error {
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		logger.Infof("Removing file %s", pidPath)
		if err := os.Remove(pidPath); err != nil {
			return err
		}
	}
	return nil
}

// Uninstall ...
func Uninstall(cfg *Config) error {
	if err := stopProcess(cfg); err != nil {
		if err != errNotRunning {
			return err
		}
	}

	appDir := cfg.AppDir()
	logger.Infof("Removing app directory %s", appDir)
	if err := os.RemoveAll(appDir); err != nil {
		return err
	}

	fmt.Printf("Uninstalled %q.\n", cfg.AppName())
	fmt.Printf("Run `keysd -reset-keyring` to remove keyring items.\n")
	return nil
}
