package git

import (
	"fmt"
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
	logger.Debugf("Get %s", id)

	if err := r.updateFilesCache(); err != nil {
		return nil, err
	}
	file, ok := r.cache.current[id]
	if !ok {
		return nil, nil
	}
	if file.deleted {
		return nil, nil
	}
	path := filepath.Join(r.Path(), r.opts.krd, file.Name())
	return ioutil.ReadFile(path) // #nosec
}

// Set data for id.
// We always save as a new file (as the next version number).
func (r *Repository) Set(id string, data []byte) error {
	if id == "" {
		return errors.Errorf("no id specified")
	}
	logger.Debugf("Set %s", id)
	now := r.opts.nowFn()

	if err := r.updateFilesCache(); err != nil {
		return err
	}

	file, ok := r.cache.current[id]
	if !ok {
		file = newFile(id, 1, now)
		logger.Debugf("Set (new) %s", file)
	} else {
		file = file.Next(now)
		logger.Debugf("Set (next) %s", file)
	}

	name := file.Name()
	msg := fmt.Sprintf("Set %s (%d)", id, file.version)
	return r.add(name, data, msg)
}

func (r *Repository) add(name string, data []byte, msg string) error {
	dir := filepath.Join(r.Path(), r.opts.krd)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path := filepath.Join(dir, name)
	if err := ioutil.WriteFile(path, data, 0600); err != nil {
		return errors.Wrapf(err, "failed to write file")
	}

	addPath := filepath.Join(r.opts.krd, name)
	if err := r.addCommit(addPath, msg); err != nil {
		// TODO: How do we resolve invalid state?
		return errors.Wrapf(err, "failed to add to repo")
	}
	return nil
}

// Delete bytes.
func (r *Repository) Delete(id string) (bool, error) {
	if id == "" {
		return false, errors.Errorf("no id specified")
	}

	if err := r.updateFilesCache(); err != nil {
		return false, err
	}

	file, ok := r.cache.current[id]
	if !ok {
		return false, nil
	}
	if file.deleted {
		return false, nil
	}

	next := file.Next(r.opts.nowFn())
	next.deleted = true
	name := next.Name()

	msg := fmt.Sprintf("Delete %s (%d)", id, next.version)
	if err := r.add(name, []byte{}, msg); err != nil {
		return false, err
	}

	return true, nil
}

// IDs ...
func (r *Repository) IDs(opts ...keyring.IDsOption) ([]string, error) {
	options := keyring.NewIDsOptions(opts...)
	prefix, showHidden, showReserved := options.Prefix, options.Hidden, options.Reserved

	if err := r.updateFilesCache(); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(r.cache.ids))
	for _, id := range r.cache.ids {
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
	if err := r.updateFilesCache(); err != nil {
		return false, err
	}
	file, ok := r.cache.current[id]
	if !ok {
		return false, nil
	}
	if file.deleted {
		return false, nil
	}
	return true, nil
}

// Reset ...
func (r *Repository) Reset() error {
	return errors.Errorf("reset not supported for git keyring")
}
