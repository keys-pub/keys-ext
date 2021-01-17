package sqlcipher

import (
	"database/sql"

	"github.com/keys-pub/keys/dstore"
)

type iterator struct {
	rows   *sql.Rows
	index  int
	limit  int
	count  int
	noData bool
}

func (i *iterator) Next() (*dstore.Document, error) {
	if !i.rows.Next() {
		// Catch auto commit errors? (This may be unnecessary in our context?)
		if err := i.rows.Close(); err != nil {
			return nil, err
		}
		return nil, i.rows.Err()
	}
	record, err := rowToRecord(i.rows)
	if err != nil {
		return nil, err
	}
	return record.Document(), nil
}

func (i *iterator) Release() {
	i.rows.Close()
}
