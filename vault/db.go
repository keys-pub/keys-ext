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
func NewDB() *DB {
	return &DB{}
}

// Name for Store.
func (d *DB) Name() string {
	return "vdb"
}

// OpenAtPath opens db located at path.
func (d *DB) OpenAtPath(path string) error {
	if d.ldb != nil {
		return errors.Errorf("already open")
	}
	if path == "" || path == "/" || path == `\` {
		return errors.Errorf("invalid path")
	}

	logger.Infof("DB at %s", path)
	ldb, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return err
	}
	d.ldb = ldb
	d.path = path
	return nil
}

// Close DB.
func (d *DB) Close() {
	if d.ldb != nil {
		d.ldb.Close()
		d.ldb = nil
	}
	d.path = ""
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

// Documents iterator.
func (d *DB) Documents(opt ...ds.DocumentsOption) (ds.DocumentIterator, error) {
	if d.ldb == nil {
		return nil, errors.Errorf("db not open")
	}
	opts := ds.NewDocumentsOptions(opt...)
	prefix := opts.Prefix
	iter := d.ldb.NewIterator(ldbutil.BytesPrefix([]byte(prefix)), nil)
	return &docsIterator{
		iter:  iter,
		index: opts.Index,
		limit: opts.Limit,
	}, nil
}

// func (d *DB) Iterator(prefix string) (iterator.Iterator, error) {
// 	if d.ldb == nil {
// 		return nil, errors.Errorf("db not open")
// 	}
// 	path := ds.Path(prefix)
// 	return d.ldb.NewIterator(ldbutil.BytesPrefix([]byte(path)), nil), nil
// }

// Exists if path exists.
func (d *DB) Exists(path string) (bool, error) {
	if d.ldb == nil {
		return false, errors.Errorf("db not open")
	}
	return d.ldb.Has([]byte(path), nil)
}
