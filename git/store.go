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
func (r *Repository) Get(service string, id string) ([]byte, error) {
	if id == "" {
		return nil, errors.Errorf("failed to get keyring item: no id specified")
	}
	if id == "." || id == ".." || strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return nil, errors.Errorf("failed to get keyring item: invalid id %q", id)
	}

	path := filepath.Join(r.Path(), service, id)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	return ioutil.ReadFile(path) // #nosec
}

// Set bytes.
func (r *Repository) Set(service string, id string, data []byte, typ string) error {
	if id == "" {
		return errors.Errorf("no id specified")
	}
	dir := filepath.Join(r.Path(), service)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, id)
	if err := ioutil.WriteFile(path, data, 0600); err != nil {
		return errors.Wrapf(err, "failed to write file")
	}

	name := filepath.Join(service, id)
	if err := r.add(name); err != nil {
		// TODO: How do we resolve invalid state?
		return errors.Wrapf(err, "failed to add to repo")
	}

	return nil
}

// Delete bytes.
func (r *Repository) Delete(service string, id string) (bool, error) {
	name := filepath.Join(service, id)
	path := filepath.Join(r.Path(), service, id)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}
	if err := os.Remove(path); err != nil {
		return false, err
	}

	if err := r.delete(name); err != nil {
		// TODO: How do we resolve invalid state?
		return false, err
	}

	return true, nil
}

// IDs ...
func (r *Repository) IDs(service string, opts *keyring.IDsOpts) ([]string, error) {
	if opts == nil {
		opts = &keyring.IDsOpts{}
	}
	prefix, showHidden, showReserved := opts.Prefix, opts.ShowHidden, opts.ShowReserved

	path := filepath.Join(r.Path(), service)
	if _, err := os.Stat(path); os.IsNotExist(err) {
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

// List items.
func (r *Repository) List(service string, key keyring.SecretKey, opts *keyring.ListOpts) ([]*keyring.Item, error) {
	return keyring.List(r, service, key, opts)
}

// Exists ...
func (r *Repository) Exists(service string, id string) (bool, error) {
	path := filepath.Join(r.Path(), service, id)
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

// Reset ...
func (r *Repository) Reset(service string) error {
	return errors.Errorf("reset not supported for git keyring")
}
