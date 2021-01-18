package sqlcipher

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys/dstore"
	sqlite3 "github.com/mutecomm/go-sqlcipher/v4"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
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

func open(path string, key SecretKey) (*sql.DB, error) {
	keyString := hex.EncodeToString(key[:])
	pragma := fmt.Sprintf("?_pragma_key=x'%s'&_pragma_cipher_page_size=4096", keyString)

	db, err := sql.Open("sqlite3", path+pragma)
	if err != nil {
		return nil, err
	}

	sqlStmt := `create table if not exists documents (
		path text not null primary key, 
		doc blob, 
		createdAt timestamp not null, 
		updatedAt timestamp not null
	);`
	if _, err = db.Exec(sqlStmt); err != nil {
		return nil, err
	}

	ok, err := sqlite3.IsEncrypted(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to check if encrypted")
	}
	if !ok {
		return nil, errors.Errorf("not encrypted")
	}

	return db, nil
}

func (d *DB) insertOrReplace(path string, r *record) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert or replace into documents (path, doc, createdAt, updatedAt) values (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	b, err := msgpack.Marshal(r.Values)
	if err != nil {
		return err
	}
	logger.Debugf("Insert/replace %s", path)
	if _, err = stmt.Exec(path, b, r.CreatedAt, r.UpdatedAt); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *DB) get(path string) (*record, error) {
	stmt, err := d.db.Prepare("select path, doc, createdAt, updatedAt from documents where path = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(path)
	return rowToRecord(row)
}

type row interface {
	Scan(dest ...interface{}) error
}

func rowToRecord(row row) (*record, error) {
	var rpath string
	var b []byte
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(&rpath, &b, &createdAt, &updatedAt); err != nil {
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

func (d *DB) iterator(ctx context.Context, parent string, opt ...dstore.Option) (*iterator, error) {
	opts := dstore.NewOptions(opt...)

	var iterPrefix string
	if parent != "/" {
		if opts.Prefix != "" {
			iterPrefix = dstore.Path(parent, opts.Prefix)
		} else {
			iterPrefix = dstore.Path(parent) + "/"
		}
	} else {
		iterPrefix = opts.Prefix
	}

	logger.Debugf("Select %s", iterPrefix)
	rows, err := d.db.Query("select path, doc, createdAt, updatedAt from documents where path like ? order by path", iterPrefix+"%")
	if err != nil {
		return nil, err
	}

	return &iterator{
		rows:   rows,
		index:  opts.Index,
		limit:  opts.Limit,
		noData: opts.NoData,
	}, nil
}

// Query ...
func (d *DB) Query(where string, args ...interface{}) (dstore.Iterator, error) {
	sql := fmt.Sprintf("select path, doc, createdAt, updatedAt from documents where %s", where)
	rows, err := d.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	return &iterator{
		rows: rows,
	}, nil
}

// Spew ...
func (d *DB) Spew(w io.Writer) error {
	rows, err := d.db.Query("select path, doc, createdAt, updatedAt from documents")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var rpath string
		var b []byte
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&rpath, &b, &createdAt, &updatedAt); err != nil {
			return err
		}
		fmt.Fprintf(w, "%s\n%s\n", rpath, spew.Sdump(b))
	}

	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
