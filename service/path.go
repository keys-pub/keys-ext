package service

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

func fileExists(p string) (bool, error) {
	if _, err := os.Stat(p); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func nextPath(p string) (string, error) {
	for i := 0; i < 1000; i++ {
		out := incPath(p, i)
		exists, err := fileExists(out)
		if err != nil {
			return "", err
		}
		if !exists {
			return out, nil
		}
	}
	return "", errors.Errorf("file already exists %s", p)
}

func incPath(p string, n int) string {
	if n == 0 {
		return p
	}
	ext := path.Ext(p)
	base := strings.TrimSuffix(p, ext)
	return fmt.Sprintf("%s-%d%s", base, n, ext)
}
