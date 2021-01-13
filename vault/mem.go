package vault

import (
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// NewMem returns an in memory Store useful for testing or ephemeral keys.
func NewMem() Store {
	return &mem{
		entries: map[string][]byte{},
	}
}

type mem struct {
	open    bool
	entries map[string][]byte
}

func (m *mem) Path() string {
	return "mem://"
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
	m.entries = map[string][]byte{}
	return nil
}

func (m *mem) Get(path string) ([]byte, error) {
	if !m.open {
		return nil, ErrNotOpen
	}
	if path == "" {
		return nil, errors.Errorf("invalid path")
	}
	if b, ok := m.entries[path]; ok {
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
	m.entries[path] = b
	return nil
}

func (m *mem) Exists(path string) (bool, error) {
	if !m.open {
		return false, ErrNotOpen
	}
	if path == "" {
		return false, errors.Errorf("invalid path")
	}
	_, ok := m.entries[path]
	return ok, nil
}

func (m *mem) Delete(path string) (bool, error) {
	if !m.open {
		return false, ErrNotOpen
	}
	if path == "" {
		return false, errors.Errorf("invalid path")
	}
	if _, ok := m.entries[path]; ok {
		delete(m.entries, path)
		return true, nil
	}
	return false, nil
}

func (m *mem) List(opts *ListOptions) ([]*Entry, error) {
	if !m.open {
		return nil, ErrNotOpen
	}
	if opts == nil {
		opts = &ListOptions{}
	}

	prefix := opts.Prefix
	out := make([]*Entry, 0, len(m.entries))
	for path, b := range m.entries {
		if strings.HasPrefix(path, prefix) {
			out = append(out, &Entry{Path: path, Data: b})
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
