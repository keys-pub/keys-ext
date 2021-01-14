package sqlcipher

import (
	"database/sql"
	"time"

	"github.com/keys-pub/keys/dstore"
	"github.com/vmihailenco/msgpack/v4"
)

type record struct {
	Path      string
	Values    map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *record) Document() *dstore.Document {
	out := &dstore.Document{
		Path:      r.Path,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	out.SetAll(r.Values)
	return out
}

// Set in DB.
func (d *DB) put(path string, r *record) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert or replace into documents (path, vals, createdAt, updatedAt) values (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	b, err := msgpack.Marshal(r.Values)
	if err != nil {
		return err
	}
	if _, err = stmt.Exec(path, b, r.CreatedAt, r.UpdatedAt); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (d *DB) get(path string) (*record, error) {
	stmt, err := d.db.Prepare("select path, vals, createdAt, updatedAt from documents where path = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var rpath string
	var b []byte
	var createdAt time.Time
	var updatedAt time.Time
	if err = stmt.QueryRow(path).Scan(&rpath, &b, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var vals map[string]interface{}
	if err := msgpack.Unmarshal(b, &vals); err != nil {
		return nil, err
	}
	return &record{
		Path:      rpath,
		Values:    vals,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (d *DB) delete(path string) (bool, error) {
	exists, err := d.exists(path)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	stmt, err := d.db.Prepare("delete from documents where path = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(path); err != nil {
		return false, err
	}
	return true, nil
}

func (d *DB) exists(path string) (bool, error) {
	stmt, err := d.db.Prepare("select 1 from documents where path = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	var value int
	if err = stmt.QueryRow(path).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return value == 1, nil
}
