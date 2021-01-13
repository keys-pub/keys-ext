package leveldb

import (
	"database/sql"
	"os"

	"github.com/keys-pub/keys-ext/vault"
	"github.com/pkg/errors"
)

var _ vault.Store = &scdb{}

type scdb struct {
	path string
	db   *sql.DB
}

// New creates sqlcipher Store.
func New(path string) vault.Store {
	return &scdb{
		path: path,
	}
}

// Path to store.
func (d *scdb) Path() string {
	return d.path
}

// Open db.
func (d *scdb) Open() error {
	if d.db != nil {
		return vault.ErrAlreadyOpen
	}
	if d.path == "" || d.path == "/" || d.path == `\` {
		return errors.Errorf("invalid path")
	}

	db, err := sql.Open("sqlite3", d.path)
	if err != nil {
		return err
	}

	sqlStmt := `create table if not exists kv (key text not null primary key, value text);`
	if _, err = db.Exec(sqlStmt); err != nil {
		return err
	}

	d.db = db
	return nil
}

// Close db.
func (d *scdb) Close() error {
	if d.db != nil {
		if err := d.db.Close(); err != nil {
			return err
		}
		d.db = nil
	}
	return nil
}

// Reset db.
func (d *scdb) Reset() error {
	wasOpen := false
	if d.db != nil {
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
func (d *scdb) Set(path string, b []byte) error {
	if d.db == nil {
		return vault.ErrNotOpen
	}
	// if err := d.ldb.Put([]byte(path), b, nil); err != nil {
	// 	return err
	// }
	return nil
}

// Get from DB.
func (d *scdb) Get(path string) ([]byte, error) {
	if d.db == nil {
		return nil, vault.ErrNotOpen
	}
	// b, err := d.scdb.Get([]byte(path), nil)
	// if err != nil {
	// 	if err == leveldb.ErrNotFound {
	// 		return nil, nil
	// 	}
	// 	return nil, err
	// }
	return nil, nil
}

// Delete from DB.
func (d *scdb) Delete(path string) (bool, error) {
	if d.db == nil {
		return false, vault.ErrNotOpen
	}
	// exists, err := d.scdb.Has([]byte(path), nil)
	// if err != nil {
	// 	return false, err
	// }
	// if !exists {
	// 	return false, nil
	// }
	// if err := d.ldb.Delete([]byte(path), nil); err != nil {
	// 	return false, err
	// }
	// return true, nil
	return false, nil
}

// List ...
func (d *scdb) List(opts *vault.ListOptions) ([]*vault.Entry, error) {
	if d.db == nil {
		return nil, vault.ErrNotOpen
	}
	if opts == nil {
		opts = &vault.ListOptions{}
	}

	// prefix := opts.Prefix
	// iter := d.ldb.NewIterator(ldbutil.BytesPrefix([]byte(prefix)), nil)
	// defer iter.Release()

	// out := []*vault.Entry{}
	// for iter.Next() {
	// 	if opts.Limit > 0 && len(out) >= opts.Limit {
	// 		break
	// 	}
	// 	path := string(iter.Key())
	// 	entry := &vault.Entry{Path: path}
	// 	if !opts.NoData {
	// 		// Remember that the contents of the returned slice should not be modified, and are
	// 		// only valid until the next call to Next.
	// 		b := copyBytes(iter.Value())
	// 		entry.Data = b
	// 	}
	// 	out = append(out, entry)

	// }
	// if err := iter.Error(); err != nil {
	// 	return nil, err
	// }
	// return out, nil
	return nil, nil
}

// Exists if path exists.
func (d *scdb) Exists(path string) (bool, error) {
	if d.db == nil {
		return false, vault.ErrNotOpen
	}
	// return d.ldb.Has([]byte(path), nil)
	return false, nil
}
