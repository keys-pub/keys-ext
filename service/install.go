package service

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

func exeDir() string {
	exe, err := ExecutablePath()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(exe)
}

func defaultBinPath() string {
	dir := exeDir()
	name := "keys"
	if runtime.GOOS == "windows" {
		name = "keys.exe"
	}
	return filepath.Join(dir, name)
}

func defaultServicePath() string {
	dir := exeDir()
	name := "keysd"
	if runtime.GOOS == "windows" {
		name = "keysd.exe"
	}
	return filepath.Join(dir, name)
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

func startFromApp(cfg *Config) error {
	if !cfg.GetBool("disableSymlinkCheck") {
		if err := installSymlink(); err != nil {
			logger.Warningf("Failed to install symlink: %s", err)
		}
		cfg.Set("disableSymlinkCheck", "1")
		if err := cfg.Save(); err != nil {
			return err
		}
	}
	return restart(cfg)
}

func installSymlink() error {
	logger.Infof("Install symlink")
	if runtime.GOOS == "windows" {
		return errors.Errorf("failed to install symlink: not implemented on windows")
	}

	binPath := defaultBinPath()
	linkDir := "/usr/local/bin"
	linkPath := filepath.Join(linkDir, "keys")

	logger.Infof("Checking if %s exists", linkDir)
	// Check if /usr/local/bin directory exists
	if _, err := os.Stat(linkDir); os.IsNotExist(err) {
		return errors.Errorf("failed to install symlink: %s does not exist", linkDir)
	}

	logger.Infof("Checking if %s exists", linkPath)
	// Check if /usr/local/bin/keys exists
	if _, err := os.Stat(linkPath); err == nil {
		logger.Infof("%s already exists", linkPath)
		return nil
	} else if os.IsNotExist(err) {
		// OK
		logger.Infof("%s doesn't exist", linkPath)
	} else {
		return err
	}

	logger.Infof("Linking %s to %s", linkPath, binPath)
	return os.Symlink(binPath, linkPath)
}
