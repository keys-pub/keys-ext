package service

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// executablePath returns path to the current executable. If the executable path
// is a symlink, the target path is returned.
func executablePath() (string, error) {
	exePath, exeErr := os.Executable()
	if exeErr != nil {
		return "", errors.Wrapf(exeErr, "os.Executable failed")
	}
	out, evalErr := filepath.EvalSymlinks(exePath)
	if evalErr != nil {
		return "", errors.Wrapf(evalErr, "eval symlinks failed")
	}
	return out, nil
}
