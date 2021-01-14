package sqlcipher

import "github.com/keys-pub/keys/dstore"

type iterator struct {
	index  int
	limit  int
	count  int
	noData bool
}

func (i *iterator) Next() (*dstore.Document, error) {
	return nil, nil
}

func (i *iterator) Release() {
}
