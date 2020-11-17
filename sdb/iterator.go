package sdb

import (
	"github.com/keys-pub/keys/dstore"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type docsIterator struct {
	db     *DB
	iter   iterator.Iterator
	index  int
	limit  int
	count  int
	noData bool
}

func (i *docsIterator) Next() (*dstore.Document, error) {
	for i.iter.Next() {
		// logger.Debugf("Document iterator path %s", path)
		i.count++
		if i.index > i.count-1 {
			continue
		}
		if i.limit != 0 && i.count > i.limit {
			return nil, nil
		}
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		path := string(i.iter.Key())
		if i.noData {
			return dstore.NewDocument(path), nil
		}
		doc, err := i.db.unmarshal(path, i.iter.Value())
		if err != nil {
			return nil, err
		}
		if doc.Path != path {
			return nil, errors.Errorf("document path mismatch")
		}
		return newDocument(doc), nil
	}
	if err := i.iter.Error(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *docsIterator) Release() {
	i.iter.Release()
}
