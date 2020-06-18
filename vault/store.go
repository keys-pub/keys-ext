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

	// Exists returns true if exists.
	Exists(path string) (bool, error)

	Documents(opt ...ds.DocumentsOption) (ds.DocumentIterator, error)
}

// Paths from vault Store.
func Paths(st Store, prefix string) ([]string, error) {
	iter, err := st.Documents(ds.Prefix(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	paths := []string{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if doc == nil {
			break
		}
		paths = append(paths, doc.Path)

	}
	return paths, nil
}

// Documents from vault Store.
func Documents(st Store, prefix string) ([]*ds.Document, error) {
	iter, err := st.Documents(ds.Prefix(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	docs := []*ds.Document{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if doc == nil {
			break
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func deleteAll(st Store, paths []string) error {
	for _, p := range paths {
		if _, err := st.Delete(p); err != nil {
			return err
		}
	}
	return nil
}
