package vault

import (
	"github.com/keys-pub/keys/ds"
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
	Documents(opt ...ds.DocumentsOption) ([]*ds.Document, error)

	// Open store.
	Open() error
	// Close store.
	Close() error
}

// Paths from vault Store.
func Paths(st Store, prefix string) ([]string, error) {
	docs, err := st.Documents(ds.Prefix(prefix))
	if err != nil {
		return nil, err
	}
	paths := []string{}
	for _, doc := range docs {
		paths = append(paths, doc.Path)

	}
	return paths, nil
}

func deleteAll(st Store, paths []string) error {
	for _, p := range paths {
		if _, err := st.Delete(p); err != nil {
			return err
		}
	}
	return nil
}

// Collections from Store.
func Collections(st Store, parent string) ([]*ds.Collection, error) {
	// We iterate over all the paths to build the collections list, this is slow.
	collections := []*ds.Collection{}
	docs, err := st.Documents(ds.Prefix(ds.Path(parent)), ds.NoData())
	if err != nil {
		return nil, err
	}
	count := map[string]int{}
	for _, doc := range docs {
		col := ds.PathFirst(doc.Path)
		colv, ok := count[col]
		if !ok {
			collections = append(collections, &ds.Collection{Path: ds.Path(col)})
			count[col] = 1
		} else {
			count[col] = colv + 1
		}
	}

	return collections, nil
}
