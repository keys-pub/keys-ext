package vault

import (
	"github.com/keys-pub/keys/docs"
)

// Store is the interface used to store data.
type Store interface {
	// Name of the Store implementation.
	Name() string

	// Get bytes.
	Get(path string) ([]byte, error)
	// Set bytes.
	Set(path string, data []byte) error
	// Delete bytes.
	Delete(path string) (bool, error)

	// Documents iterator.
	Documents(opt ...docs.Option) ([]*docs.Document, error)

	// Open store.
	Open() error
	// Close store.
	Close() error
}

func deleteAll(st Store, paths []string) error {
	for _, p := range paths {
		if _, err := st.Delete(p); err != nil {
			return err
		}
	}
	return nil
}

// Collections from vault db.
func (v *Vault) Collections(parent string) ([]*docs.Collection, error) {
	// We iterate over all the paths to build the collections list, this is slow.
	collections := []*docs.Collection{}
	ds, err := v.store.Documents(docs.Prefix(docs.Path(parent)), docs.NoData())
	if err != nil {
		return nil, err
	}
	count := map[string]int{}
	for _, doc := range ds {
		col := docs.PathFirst(doc.Path)
		colv, ok := count[col]
		if !ok {
			collections = append(collections, &docs.Collection{Path: docs.Path(col)})
			count[col] = 1
		} else {
			count[col] = colv + 1
		}
	}

	return collections, nil
}

// Documents from Store.
func (v *Vault) Documents(opt ...docs.Option) ([]*docs.Document, error) {
	return v.store.Documents(opt...)
}

// DeleteDocument remotes document from vault.
func (v *Vault) DeleteDocument(path string) (bool, error) {
	return v.store.Delete(path)
}
