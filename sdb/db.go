package sdb

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/vmihailenco/msgpack/v4"
)

var _ ds.DocumentStore = &DB{}

// SecretKey for database.
type SecretKey *[32]byte

// DB is leveldb implementation of ds.DocumentStore.
type DB struct {
	rwmtx *sync.RWMutex
	sdb   *sdb
	path  string
	clock func() time.Time

	inc    int
	incMax int
}

// New creates a DB.
func New() *DB {
	return &DB{
		rwmtx: &sync.RWMutex{},
		clock: time.Now,
	}
}

// SetClock sets clock.
func (d *DB) SetClock(clock func() time.Time) {
	d.clock = clock
}

// OpenAtPath opens db located at path.
func (d *DB) OpenAtPath(ctx context.Context, path string, key SecretKey) error {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.sdb != nil {
		return errors.Errorf("already open")
	}

	logger.Infof("Open %s", path)
	d.path = path
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return err
	}
	sdb := newSDB(db, key)
	d.sdb = sdb
	return nil
}

// Close the db.
func (d *DB) Close() {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()

	if d.sdb != nil {
		logger.Infof("Closing db %s", d.path)
		if err := d.sdb.Close(); err != nil {
			logger.Errorf("Error closing DB: %s", err)
		}
		d.sdb = nil
	}
}

// Exists returns true if the db row exists at path
func (d *DB) Exists(ctx context.Context, path string) (bool, error) {
	if d.sdb == nil {
		return false, errors.Errorf("db not open")
	}
	path = ds.Path(path)
	return d.sdb.Has(path)
}

// Create entry.
func (d *DB) Create(ctx context.Context, path string, b []byte) error {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.sdb == nil {
		return errors.Errorf("db not open")
	}
	path = ds.Path(path)
	if path == "/" {
		return errors.Errorf("invalid path %s", path)
	}
	exists, err := d.Exists(ctx, path)
	if err != nil {
		return err
	}
	if exists {
		return ds.NewErrPathExists(path)
	}

	now := d.clock()
	doc := &ds.Document{
		Path:      path,
		Data:      b,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mb, err := msgpack.Marshal(doc)
	if err != nil {
		return err
	}

	if err := d.sdb.Put(path, mb); err != nil {
		return err
	}

	return nil
}

// Set saves document to the db at key.
func (d *DB) Set(ctx context.Context, path string, b []byte) error {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.sdb == nil {
		return errors.Errorf("db not open")
	}
	path = ds.Path(path)
	if path == "/" {
		return errors.Errorf("invalid path %s", path)
	}
	return d.set(ctx, path, b)
}

func (d *DB) set(ctx context.Context, path string, b []byte) error {
	doc, err := d.Get(ctx, path)
	if err != nil {
		return err
	}
	now := d.clock()
	if doc == nil {
		doc = &ds.Document{
			Path:      path,
			Data:      b,
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else {
		doc.UpdatedAt = now
		doc.Data = b
	}

	mb, err := msgpack.Marshal(doc)
	if err != nil {
		return err
	}

	if err := d.sdb.Put(path, mb); err != nil {
		return err
	}

	return nil
}

// Get entry at path.
func (d *DB) Get(ctx context.Context, path string) (*ds.Document, error) {
	path = ds.Path(path)
	doc, err := d.get(ctx, path)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	return doc, nil
}

// GetAll paths.
func (d *DB) GetAll(ctx context.Context, paths []string) ([]*ds.Document, error) {
	out := make([]*ds.Document, 0, len(paths))
	for _, p := range paths {
		// TODO: Handle context Done()
		doc, err := d.get(ctx, p)
		if err != nil {
			return nil, err
		}
		if doc == nil {
			continue
		}
		out = append(out, doc)
	}
	return out, nil
}

// Collections ...
func (d *DB) Collections(ctx context.Context, parent string) ([]*ds.Collection, error) {
	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}
	if ds.Path(parent) != "/" {
		// TODO: Support nested collections
		return nil, errors.Errorf("only root collections supported")
	}

	// We iterate over all the paths to build the collections list, this is slow.
	collections := []*ds.Collection{}
	count := map[string]int{}
	iter := &docsIterator{
		db:   d,
		iter: d.sdb.NewIterator(""),
	}
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if doc == nil {
			break
		}
		col := ds.PathFirst(doc.Path)
		colv, ok := count[col]
		if !ok {
			collections = append(collections, &ds.Collection{Path: ds.Path(col)})
			count[col] = 1
		} else {
			count[col] = colv + 1
		}
	}
	return collections, nil
}

// Delete value at path.
func (d *DB) Delete(ctx context.Context, path string) (bool, error) {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.sdb == nil {
		return false, errors.Errorf("db not open")
	}
	path = ds.Path(path)
	ok, err := d.sdb.Has(path)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	logger.Infof("Deleting %s", path)
	if err := d.sdb.Delete(path); err != nil {
		return false, err
	}
	return true, nil
}

// DeleteAll paths.
func (d *DB) DeleteAll(ctx context.Context, paths []string) error {
	for _, p := range paths {
		if err := d.sdb.Delete(p); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) document(path string, b []byte) (*ds.Document, error) {
	var doc ds.Document
	if err := msgpack.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	if doc.Path != path {
		return nil, errors.Errorf("document path mismatch %s != %s", doc.Path, path)
	}
	return &doc, nil
}

// DocumentIterator ...
func (d *DB) DocumentIterator(ctx context.Context, parent string, opt ...ds.DocumentsOption) (ds.DocumentIterator, error) {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()

	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}

	opts := ds.NewDocumentsOptions(opt...)

	iter, err := d.iterator(ctx, parent, opts.Prefix)
	if err != nil {
		return nil, err
	}

	// TODO: Handle context Done()
	return &docsIterator{
		db:    d,
		iter:  iter,
		index: opts.Index,
		limit: opts.Limit,
	}, nil
}

func (d *DB) iterator(ctx context.Context, parent string, prefix string) (iterator.Iterator, error) {
	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}

	var iterPrefix string
	if parent != "" {
		if prefix != "" {
			iterPrefix = ds.Path(parent, prefix)
		} else {
			iterPrefix = ds.Path(parent) + "/"
		}
	} else {
		iterPrefix = prefix
	}

	// TODO: Handle context Done()
	return d.sdb.NewIterator(iterPrefix), nil
}

// Documents ...
func (d *DB) Documents(ctx context.Context, parent string, opt ...ds.DocumentsOption) ([]*ds.Document, error) {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()
	opts := ds.NewDocumentsOptions(opt...)

	if opts.Index != 0 {
		return nil, errors.Errorf("index not implemented")
	}

	iter, err := d.iterator(ctx, parent, opts.Prefix)
	if err != nil {
		return nil, err
	}

	docs := []*ds.Document{}
	for iter.Next() {
		if opts.Limit > 0 && len(docs) >= opts.Limit {
			break
		}
		path := string(iter.Key())
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		doc, err := d.document(path, iter.Value())
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
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

func (d *DB) get(ctx context.Context, path string) (*ds.Document, error) {
	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}
	b, err := d.sdb.Get(path)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}

	return d.document(path, b)
}

// Last returns last item with key prefix.
func (d *DB) Last(ctx context.Context, prefix string) (*ds.Document, error) {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()
	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}
	var doc *ds.Document
	iter := d.sdb.NewIterator(prefix)
	defer iter.Release()
	if ok := iter.Last(); ok {
		path := string(iter.Value())
		val, err := d.get(ctx, path)
		if err != nil {
			return nil, err
		}
		doc = val
	}
	if err := iter.Error(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate db")
	}
	return doc, nil
}

// Count returns number of docs in a collection with prefix and filter.
// This iterates over the prefixed docs to count them.
func (d *DB) Count(ctx context.Context, prefix string, contains string) (int, error) {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()
	return d.countEntries(prefix, contains)
}

func (d *DB) countEntries(prefix string, contains string) (int, error) {
	if d.sdb == nil {
		return 0, errors.Errorf("db not open")
	}
	var prefixRange string
	if prefix != "" {
		prefixRange = prefix
	}
	iter := d.sdb.NewIterator(prefixRange)
	defer iter.Release()
	total := 0
	for iter.Next() {
		path := string(iter.Key())
		if contains != "" {
			value := iter.Value()
			entry := ds.NewDocument(path, value)
			if entry.Contains(contains) {
				total++
			}
		} else {
			total++
		}
	}
	if err := iter.Error(); err != nil {
		return -1, err
	}
	return total, nil
}

// Spew ...
func (d *DB) Spew(prefix string, out io.Writer) error {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()
	if d.sdb == nil {
		return errors.Errorf("db not open")
	}

	iter := d.sdb.NewIterator(prefix)
	defer iter.Release()
	for iter.Next() {
		k := string(iter.Key())
		s := fmt.Sprintf("%s %s\n", k, spew.Sdump(iter.Value()))
		if _, err := out.Write([]byte(s)); err != nil {
			return err
		}
	}
	if err := iter.Error(); err != nil {
		return err
	}
	return nil
}
