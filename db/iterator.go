package db

import (
	"github.com/keys-pub/keys/ds"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type docsIterator struct {
	db    *DB
	iter  iterator.Iterator
	index int
	limit int
	count int
}

func (i *docsIterator) Next() (*ds.Document, error) {
	for i.iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		path := string(i.iter.Key())
		logger.Debugf("Document iterator path %s", path)
		i.count++
		if i.index > i.count-1 {
			continue
		}
		if i.limit != 0 && i.count > i.limit {
			return nil, nil
		}
		return i.db.document(path, i.iter.Value())
	}
	if err := i.iter.Error(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *docsIterator) Release() {
	i.iter.Release()
}

type colsIterator struct {
	iter iterator.Iterator
}

func (i *colsIterator) Next() (*ds.Collection, error) {
	for i.iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		path := string(i.iter.Key())

		logger.Debugf("Collection iterator path %s", path)
		return &ds.Collection{Path: ds.Path(path[1:])}, nil
	}
	if err := i.iter.Error(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *colsIterator) Release() {
	i.iter.Release()
}
