package vault

import (
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	ldbutil "github.com/syndtr/goleveldb/leveldb/util"
)

var _ Store = &DB{}

// DB Store.
type DB struct {
	ldb  *leveldb.DB
	path string
}

// NewDB creates DB Store.
func NewDB(path string) *DB {
	return &DB{
		path: path,
	}
}

// Name for Store.
func (d *DB) Name() string {
	return "vdb"
}

// Open db.
func (d *DB) Open() error {
	if d.ldb != nil {
		return errors.Errorf("already open")
	}
	if d.path == "" || d.path == "/" || d.path == `\` {
		return errors.Errorf("invalid path")
	}

	logger.Infof("Open %s", d.path)
	ldb, err := leveldb.OpenFile(d.path, nil)
	if err != nil {
		return err
	}
	d.ldb = ldb
	return nil
}

// Close db.
func (d *DB) Close() error {
	if d.ldb != nil {
		if err := d.ldb.Close(); err != nil {
			return err
		}
		d.ldb = nil
	}
	return nil
}

// Set in DB.
func (d *DB) Set(path string, b []byte) error {
	if d.ldb == nil {
		return errors.Errorf("db not open")
	}
	if err := d.ldb.Put([]byte(path), b, nil); err != nil {
		return err
	}
	return nil
}

// Get from DB.
func (d *DB) Get(path string) ([]byte, error) {
	if d.ldb == nil {
		return nil, errors.Errorf("db not open")
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
func (d *DB) Delete(path string) (bool, error) {
	if d.ldb == nil {
		return false, errors.Errorf("db not open")
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

// Documents ...
func (d *DB) Documents(opt ...ds.DocumentsOption) ([]*ds.Document, error) {
	if d.ldb == nil {
		return nil, errors.Errorf("db not open")
	}
	opts := ds.NewDocumentsOptions(opt...)

	if opts.Index != 0 {
		return nil, errors.Errorf("index not implemented")
	}

	prefix := opts.Prefix
	iter := d.ldb.NewIterator(ldbutil.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	docs := []*ds.Document{}
	for iter.Next() {
		if opts.Limit > 0 && len(docs) >= opts.Limit {
			break
		}
		path := string(iter.Key())
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		b := copyBytes(iter.Value())
		docs = append(docs, ds.NewDocument(path, b))
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return docs, nil
}

func copyBytes(source []byte) []byte {
	dest := make([]byte, len(source))
	copy(dest, source)
	return dest
}

// Exists if path exists.
func (d *DB) Exists(path string) (bool, error) {
	if d.ldb == nil {
		return false, errors.Errorf("db not open")
	}
	return d.ldb.Has([]byte(path), nil)
}
