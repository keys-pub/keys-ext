package vault

import (
	"strings"

	"github.com/keys-pub/keys/dstore"
)

// Entry in Store.
type Entry struct {
	Path string
	Data []byte
}

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

	// List store entries.
	List(opts *ListOptions) ([]*Entry, error)

	// Open store.
	Open() error
	// Close store.
	Close() error
	// Reset store.
	Reset() error
}

// ListOptions for listing Store.
type ListOptions struct {
	Prefix string
	NoData bool
	Limit  int
}

func deleteAll(st Store, paths []string) error {
	for _, p := range paths {
		if _, err := st.Delete(p); err != nil {
			return err
		}
	}
	return nil
}

// Collections lists collection paths from parent.
func Collections(st Store, parent string) ([]string, error) {
	entries, err := st.List(&ListOptions{Prefix: parent, NoData: true})
	if err != nil {
		return nil, err
	}
	out := []string{}
	cols := map[string]bool{}
	for _, entry := range entries {
		col := dstore.Path(dstore.PathFirst(strings.TrimPrefix(entry.Path, parent)))
		_, ok := cols[col]
		if ok {
			continue
		}
		cols[col] = true
		out = append(out, col)
	}
	return out, nil
}
