package vault

import (
	"sort"
	"strings"

	"github.com/keys-pub/keys/dstore"
	"github.com/pkg/errors"
)

// NewMem returns an in memory Store useful for testing or ephemeral keys.
func NewMem() Store {
	return &mem{
		items: map[string][]byte{},
	}
}

type mem struct {
	open  bool
	items map[string][]byte
}

func (m *mem) Name() string {
	return "mem"
}

func (m *mem) Open() error {
	if m.open {
		return ErrAlreadyOpen
	}
	m.open = true
	return nil
}

func (m *mem) Close() error {
	m.open = false
	return nil
}

func (m *mem) Reset() error {
	m.items = map[string][]byte{}
	return nil
}

func (m *mem) Get(path string) ([]byte, error) {
	if !m.open {
		return nil, ErrNotOpen
	}
	if path == "" {
		return nil, errors.Errorf("invalid path")
	}
	if b, ok := m.items[path]; ok {
		return b, nil
	}
	return nil, nil
}

func (m *mem) Set(path string, b []byte) error {
	if !m.open {
		return ErrNotOpen
	}
	if path == "" {
		return errors.Errorf("invalid path")
	}
	m.items[path] = b
	return nil
}

func (m *mem) Exists(path string) (bool, error) {
	if !m.open {
		return false, ErrNotOpen
	}
	if path == "" {
		return false, errors.Errorf("invalid path")
	}
	_, ok := m.items[path]
	return ok, nil
}

func (m *mem) Delete(path string) (bool, error) {
	if !m.open {
		return false, ErrNotOpen
	}
	if path == "" {
		return false, errors.Errorf("invalid path")
	}
	if _, ok := m.items[path]; ok {
		delete(m.items, path)
		return true, nil
	}
	return false, nil
}

func (m *mem) Documents(opt ...dstore.Option) ([]*dstore.Document, error) {
	if !m.open {
		return nil, ErrNotOpen
	}
	opts := dstore.NewOptions(opt...)
	prefix := opts.Prefix

	out := make([]*dstore.Document, 0, len(m.items))
	for path, b := range m.items {
		if strings.HasPrefix(path, prefix) {
			out = append(out, dstore.NewDocument(path).WithData(b))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	if opts.Limit > 0 && len(out) > opts.Limit {
		out = out[:opts.Limit]
	}
	return out, nil
}
