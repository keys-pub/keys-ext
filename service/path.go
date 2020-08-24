package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const filePerms = 0600

func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func nextPathIfExists(p string) (string, error) {
	for i := 0; i < 1000; i++ {
		out := incPath(p, i)
		exists, err := pathExists(out)
		if err != nil {
			return "", err
		}
		if !exists {
			return out, nil
		}
	}
	return "", errors.Errorf("file exists %s", p)
}

func incPath(p string, n int) string {
	if n == 0 {
		return p
	}
	ext := path.Ext(p)
	base := strings.TrimSuffix(p, ext)
	return fmt.Sprintf("%s-%d%s", base, n+1, ext)
}

func isDir(path string) (bool, error) {
	if path == "" {
		return false, nil
	}
	exists, err := pathExists(path)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, err
	}
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		return true, nil
	case mode.IsRegular():
		return false, nil
	default:
		return false, nil
	}
}

// resolveOutPath returns file to output given in, out and suffixes.
//
// ("", "/dir/file.enc", ".enc") => "/dir/file"
// ("/dir/file2", "/dir/file.enc", "") => "/dir/file2"
//
func resolveOutPath(out string, in string, inSuffix string) (string, error) {
	dir, err := isDir(out)
	if err != nil {
		return "", err
	}
	if dir {
		_, inFile := filepath.Split(in)
		out = filepath.Join(out, inFile)
	}

	if out == "" {
		out = in
	}

	out = strings.TrimSuffix(out, inSuffix)

	if !filepath.IsAbs(out) {
		inDir, _ := filepath.Split(in)
		out = filepath.Join(inDir, out)
	}

	exists, err := pathExists(out)
	if err != nil {
		return "", err
	}
	if exists {
		next, err := nextPathIfExists(out)
		if err != nil {
			return "", err
		}
		out = next
	}
	return out, nil
}
