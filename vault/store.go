package vault

import "github.com/keys-pub/keys/dstore"

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
	Documents(opt ...dstore.Option) ([]*dstore.Document, error)

	// Open store.
	Open() error
	// Close store.
	Close() error
	// Reset store.
	Reset() error
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
func (v *Vault) Collections(parent string) ([]*dstore.Collection, error) {
	// We iterate over all the paths to build the collections list, this is slow.
	collections := []*dstore.Collection{}
	ds, err := v.store.Documents(dstore.Prefix(dstore.Path(parent)), dstore.NoData())
	if err != nil {
		return nil, err
	}
	count := map[string]int{}
	for _, doc := range ds {
		col := dstore.PathFirst(doc.Path)
		colv, ok := count[col]
		if !ok {
			collections = append(collections, &dstore.Collection{Path: dstore.Path(col)})
			count[col] = 1
		} else {
			count[col] = colv + 1
		}
	}

	return collections, nil
}

// Documents from Store.
func (v *Vault) Documents(opt ...dstore.Option) ([]*dstore.Document, error) {
	return v.store.Documents(opt...)
}
