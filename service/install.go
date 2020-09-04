package service

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	kenv "github.com/keys-pub/keys/env"
	"github.com/pkg/errors"
)

func exeDir() string {
	exe, err := executablePath()
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

func restart(env *Env) error {
	logger.Debugf("Restart process")
	if err := stop(env); err != nil {
		if err != errNotRunning {
			return err
		}
	}
	return autostart(env)
}

func start(env *Env, wait bool) error {
	if err := startProcess(env); err != nil {
		return err
	}
	if wait {
		if err := waitForStart(env); err != nil {
			return err
		}
	}
	return nil
}

func stop(env *Env) error {
	// TODO: This stops first process with keysd name
	if err := stopProcess(env); err != nil {
		return err
	}
	pidPath, err := env.AppPath("pid", false)
	if err != nil {
		return err
	}
	if err := removeFile(pidPath); err != nil {
		return err
	}

	return nil
}

func removeFile(pidPath string) error {
	exists, err := pathExists(pidPath)
	if err != nil {
		return err
	}
	if exists {
		logger.Infof("Removing file %s", pidPath)
		if err := os.Remove(pidPath); err != nil {
			return err
		}
	}
	return nil
}

// Uninstall ...
func Uninstall(out io.Writer, env *Env) error {
	if err := stopProcess(env); err != nil {
		if err != errNotRunning {
			return err
		}
	}

	dirs, err := kenv.AllDirs(env.AppName())
	if err != nil {
		return err
	}
	for _, d := range dirs {
		fmt.Fprintf(out, "Removing \"%s\".\n", d)
		if err := os.RemoveAll(d); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Uninstalled %q.\n", env.AppName())
	return nil
}

func startFromApp(env *Env) error {
	// TODO: Check/fix symlink if busted
	if env.GetInt("disableSymlinkCheck", 0) < 2 {
		if err := installSymlink(); err != nil {
			logger.Infof("Failed to install symlink: %s", err)
		} else {
			// Only install once
			env.Set("disableSymlinkCheck", "2")
			if err := env.Save(); err != nil {
				return err
			}
		}
	}
	return restart(env)
}

func installSymlink() error {
	logger.Infof("Install symlink")
	if runtime.GOOS == "windows" {
		return errors.Errorf("not implemented on windows")
	}

	binPath := defaultBinPath()

	if strings.HasPrefix(binPath, "/Volumes/") {
		return errors.Errorf("currently running from Volumes")
	}

	linkDir := "/usr/local/bin"
	linkPath := filepath.Join(linkDir, "keys")

	logger.Infof("Checking if %s exists", linkDir)
	linkDirExists, err := pathExists(linkDir)
	if err != nil {
		return err
	}
	// Check if /usr/local/bin directory exists
	if !linkDirExists {
		return errors.Errorf("%s does not exist", linkDir)
	}

	logger.Infof("Checking if %s exists", linkPath)
	linkExists, err := pathExists(linkPath)
	if err != nil {
		return err
	}
	// Check if /usr/local/bin/keys exists
	if linkExists {
		logger.Infof("%s already exists", linkPath)
		return nil
	} else if os.IsNotExist(err) {
		// OK
		logger.Infof("%s doesn't exist", linkPath)
	}

	logger.Infof("Linking %s to %s", linkPath, binPath)
	return os.Symlink(binPath, linkPath)
}

func checkForAppConflict() error {
	path, err := executablePath()
	if err != nil {
		return err
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}

	var check []string
	switch runtime.GOOS {
	case "darwin":
		check = []string{"/Applications/Keys.app", filepath.Join(usr.HomeDir, "Applications", "Keys.app")}
	case "windows":
		// TODO
	}
	for _, c := range check {
		if !strings.HasPrefix(path, c) {
			exists, err := pathExists(c)
			if err != nil {
				return err
			}
			if exists {
				return errors.Errorf("You have the app installed (%s), but this (%s) doesn't point there. You may have multiple installations?", c, path)
			}
		}
	}

	return nil
}
