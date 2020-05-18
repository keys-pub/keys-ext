package git

import (
	"io/ioutil"
	"path/filepath"

	"github.com/keys-pub/keys/keyring"
)

// List items.
func (r *Repository) List() ([]*keyring.Item, error) {
	logger.Debugf("List path: %s", r.path)
	files, err := ioutil.ReadDir(r.path)
	if err != nil {
		return nil, err
	}

	items := make([]*keyring.Item, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		item, err := r.get(filepath.Join(r.path, file.Name()))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// Get item.
func (r *Repository) Get(id string) (*keyring.Item, error) {
	path := filepath.Join(r.path, id)
	return r.get(path)
}

func (r *Repository) get(path string) (*keyring.Item, error) {
	encrypted, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	item, err := decryptItem(encrypted, r.key, r.ks)
	if err != nil {
		return nil, err
	}
	return item, nil
}
