package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

var _ keyring.Store = &Repository{}

// Name of git keyring.Store.
func (r *Repository) Name() string {
	return "git"
}

// Get bytes.
func (r *Repository) Get(id string) ([]byte, error) {
	if id == "" {
		return nil, errors.Errorf("failed to get keyring item: no id specified")
	}
	if id == "." || id == ".." || strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return nil, errors.Errorf("failed to get keyring item: invalid id %q", id)
	}

	path := filepath.Join(r.Path(), r.krd, id)
	exists, err := pathExists(path)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	return ioutil.ReadFile(path) // #nosec
}

// Set bytes.
func (r *Repository) Set(id string, data []byte) error {
	if id == "" {
		return errors.Errorf("no id specified")
	}
	dir := filepath.Join(r.Path(), r.krd)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, id)
	if err := ioutil.WriteFile(path, data, 0600); err != nil {
		return errors.Wrapf(err, "failed to write file")
	}

	name := filepath.Join(r.krd, id)
	if err := r.Add(name); err != nil {
		// TODO: How do we resolve invalid state?
		return errors.Wrapf(err, "failed to add to repo")
	}

	return nil
}

// Delete bytes.
func (r *Repository) Delete(id string) (bool, error) {
	name := filepath.Join(r.krd, id)
	path := filepath.Join(r.Path(), r.krd, id)
	exists, err := pathExists(path)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	if err := os.Remove(path); err != nil {
		return false, err
	}

	if err := r.Remove(name); err != nil {
		// TODO: How do we resolve invalid state?
		// TODO: Move file into tmp and the remove if successful from git rm
		return false, err
	}
	return true, nil
}

// IDs ...
func (r *Repository) IDs(opts ...keyring.IDsOption) ([]string, error) {
	options := keyring.NewIDsOptions(opts...)
	prefix, showHidden, showReserved := options.Prefix, options.Hidden, options.Reserved

	path := filepath.Join(r.Path(), r.krd)
	exists, err := pathExists(path)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{}, nil
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(files))
	for _, f := range files {
		id := f.Name()
		if !showReserved && strings.HasPrefix(id, keyring.ReservedPrefix) {
			continue
		}
		if !showHidden && strings.HasPrefix(id, keyring.HiddenPrefix) {
			continue
		}
		if prefix != "" && !strings.HasPrefix(id, prefix) {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Exists ...
func (r *Repository) Exists(id string) (bool, error) {
	path := filepath.Join(r.Path(), r.krd, id)
	return pathExists(path)
}

// Reset ...
func (r *Repository) Reset() error {
	return errors.Errorf("reset not supported for git keyring")
}
