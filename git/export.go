package git

import (
	"crypto/subtle"

	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

// Changes for items.
type Changes struct {
	Add    []*keyring.Item
	Update []*keyring.Item
}

// Import items from repository.
func Import(kr *keyring.Keyring, repo *Repository) (*Changes, error) {
	items, err := repo.List()
	if err != nil {
		return nil, err
	}

	added := []*keyring.Item{}
	updated := []*keyring.Item{}

	for _, item := range items {
		existing, err := kr.Get(item.ID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			chg, err := itemChanged(existing, item)
			if err != nil {
				return nil, err
			}
			if chg {
				updated = append(updated, item)
				if err := kr.Update(item.ID, item.Data); err != nil {
					return nil, err
				}
			}
		} else {
			// Added
			if err := kr.Create(item); err != nil {
				return nil, err
			}
			added = append(added, item)
		}
	}

	return &Changes{
		Add:    added,
		Update: updated,
	}, nil
}

// Export (changed) keyring items to repository.
func Export(kr *keyring.Keyring, repo *Repository) (*Changes, error) {
	items, err := kr.List(nil)
	if err != nil {
		return nil, err
	}

	added := []*keyring.Item{}
	updated := []*keyring.Item{}

	for _, item := range items {
		repoItem, err := repo.Get(item.ID)
		if err != nil {
			return nil, err
		}
		if repoItem != nil {
			chg, err := itemChanged(item, repoItem)
			if err != nil {
				return nil, err
			}
			if chg {
				if err := repo.Add(item); err != nil {
					return nil, err
				}
				updated = append(updated, item)
			}
		} else {
			// Added
			if err := repo.Add(item); err != nil {
				return nil, err
			}
			added = append(added, item)
		}
	}

	return &Changes{
		Add:    added,
		Update: updated,
	}, nil
}

func itemChanged(item1 *keyring.Item, item2 *keyring.Item) (bool, error) {
	if item1.ID != item2.ID {
		return false, errors.Errorf("mismatched item ids")
	}
	if item1.Type != item2.Type {
		return false, errors.Errorf("mismatched item types")
	}
	if subtle.ConstantTimeCompare(item1.Data, item2.Data) != 1 {
		return true, nil
	}
	return false, nil
}
