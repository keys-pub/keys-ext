package vault

import (
	"sort"
	"strings"

	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
)

// NewMem returns an in memory Store useful for testing or ephemeral keys.
func NewMem() Store {
	return &mem{
		items: map[string][]byte{},
	}
}

type mem struct {
	items map[string][]byte
}

func (k *mem) Name() string {
	return "mem"
}

func (k *mem) Get(path string) ([]byte, error) {
	if path == "" {
		return nil, errors.Errorf("invalid path")
	}
	if b, ok := k.items[path]; ok {
		return b, nil
	}
	return nil, nil
}

func (k *mem) Set(path string, data []byte) error {
	if path == "" {
		return errors.Errorf("invalid path")
	}
	k.items[path] = data
	return nil
}

func (k *mem) Exists(path string) (bool, error) {
	if path == "" {
		return false, errors.Errorf("invalid path")
	}
	_, ok := k.items[path]
	return ok, nil
}

func (k *mem) Delete(path string) (bool, error) {
	if path == "" {
		return false, errors.Errorf("invalid path")
	}
	if _, ok := k.items[path]; ok {
		delete(k.items, path)
		return true, nil
	}
	return false, nil
}

func (k *mem) Documents(opt ...ds.DocumentsOption) (ds.DocumentIterator, error) {
	opts := ds.NewDocumentsOptions(opt...)
	prefix := opts.Prefix

	docs := make([]*ds.Document, 0, len(k.items))
	for path, b := range k.items {
		if strings.HasPrefix(path, prefix) {
			docs = append(docs, &ds.Document{Path: path, Data: b})
		}
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Path < docs[j].Path
	})
	return ds.NewDocumentIterator(docs...), nil
}
