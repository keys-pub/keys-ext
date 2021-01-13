package leveldb

import (
	"os"

	"github.com/keys-pub/keys-ext/vault"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	ldbutil "github.com/syndtr/goleveldb/leveldb/util"
)

var _ vault.Store = &ldb{}

type ldb struct {
	ldb  *leveldb.DB
	path string
}

// New creates leveldb Store.
func New(path string) vault.Store {
	return &ldb{
		path: path,
	}
}

// Path to store.
func (d *ldb) Path() string {
	return d.path
}

// Open db.
func (d *ldb) Open() error {
	if d.ldb != nil {
		return vault.ErrAlreadyOpen
	}
	if d.path == "" || d.path == "/" || d.path == `\` {
		return errors.Errorf("invalid path")
	}

	ldb, err := leveldb.OpenFile(d.path, nil)
	if err != nil {
		return err
	}
	d.ldb = ldb
	return nil
}

// Close db.
func (d *ldb) Close() error {
	if d.ldb != nil {
		if err := d.ldb.Close(); err != nil {
			return err
		}
		d.ldb = nil
	}
	return nil
}

// Reset db.
func (d *ldb) Reset() error {
	wasOpen := false
	if d.ldb != nil {
		wasOpen = true
		if err := d.Close(); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(d.path); err != nil {
		return err
	}
	if wasOpen {
		if err := d.Open(); err != nil {
			return err
		}
	}
	return nil
}

// Set in DB.
func (d *ldb) Set(path string, b []byte) error {
	if d.ldb == nil {
		return vault.ErrNotOpen
	}
	if err := d.ldb.Put([]byte(path), b, nil); err != nil {
		return err
	}
	return nil
}

// Get from DB.
func (d *ldb) Get(path string) ([]byte, error) {
	if d.ldb == nil {
		return nil, vault.ErrNotOpen
	}
	b, err := d.ldb.Get([]byte(path), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

// Delete from DB.
func (d *ldb) Delete(path string) (bool, error) {
	if d.ldb == nil {
		return false, vault.ErrNotOpen
	}
	exists, err := d.ldb.Has([]byte(path), nil)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	if err := d.ldb.Delete([]byte(path), nil); err != nil {
		return false, err
	}
	return true, nil
}

// List ...
func (d *ldb) List(opts *vault.ListOptions) ([]*vault.Entry, error) {
	if d.ldb == nil {
		return nil, vault.ErrNotOpen
	}
	if opts == nil {
		opts = &vault.ListOptions{}
	}

	prefix := opts.Prefix
	iter := d.ldb.NewIterator(ldbutil.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	out := []*vault.Entry{}
	for iter.Next() {
		if opts.Limit > 0 && len(out) >= opts.Limit {
			break
		}
		path := string(iter.Key())
		entry := &vault.Entry{Path: path}
		if !opts.NoData {
			// Remember that the contents of the returned slice should not be modified, and are
			// only valid until the next call to Next.
			b := copyBytes(iter.Value())
			entry.Data = b
		}
		out = append(out, entry)

	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return out, nil
}

func copyBytes(source []byte) []byte {
	dest := make([]byte, len(source))
	copy(dest, source)
	return dest
}

// Exists if path exists.
func (d *ldb) Exists(path string) (bool, error) {
	if d.ldb == nil {
		return false, vault.ErrNotOpen
	}
	return d.ldb.Has([]byte(path), nil)
}
