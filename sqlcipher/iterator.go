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
	for i.rows.Next() {
		i.count++
		if i.index > i.count-1 {
			continue
		}
		if i.limit != 0 && i.count > i.limit {
			return nil, nil
		}

		record, err := rowToRecord(i.rows)
		if err != nil {
			return nil, err
		}
		return record.Document(), nil
	}
	// Catch auto commit errors? (This may be unnecessary in our context?)
	if err := i.rows.Close(); err != nil {
		return nil, err
	}
	return nil, i.rows.Err()
}

func (i *iterator) Release() {
	_ = i.rows.Close()
}
