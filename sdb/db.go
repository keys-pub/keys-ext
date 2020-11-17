package sdb

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/vmihailenco/msgpack/v4"
)

var _ dstore.Documents = &DB{}

// SecretKey for database.
type SecretKey *[32]byte

// DB is secure leveldb implementation of dstore.Documents.
type DB struct {
	rwmtx *sync.RWMutex
	sdb   *sdb
	path  string
	clock tsutil.Clock
}

// New creates a DB.
func New() *DB {
	return &DB{
		rwmtx: &sync.RWMutex{},
		clock: tsutil.NewClock(),
	}
}

// SetClock sets clock.
func (d *DB) SetClock(clock tsutil.Clock) {
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

// IsOpen returns true if open.
func (d *DB) IsOpen() bool {
	return d.sdb != nil
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
	path = dstore.Path(path)
	return d.sdb.Has(path)
}

type document struct {
	Path      string                 `msgpack:"path"`
	Values    map[string]interface{} `msgpack:"v"`
	CreatedAt time.Time              `msgpack:"cts"`
	UpdatedAt time.Time              `msgpack:"uts"`
}

// Create entry.
func (d *DB) Create(ctx context.Context, path string, values map[string]interface{}) error {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.sdb == nil {
		return errors.Errorf("db not open")
	}
	path = dstore.Path(path)
	if path == "/" {
		return errors.Errorf("invalid path %s", path)
	}
	exists, err := d.Exists(ctx, path)
	if err != nil {
		return err
	}
	if exists {
		return dstore.NewErrPathExists(path)
	}

	now := d.clock.Now()
	doc := &document{
		Path:      path,
		Values:    values,
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
func (d *DB) Set(ctx context.Context, path string, values map[string]interface{}, opt ...dstore.SetOption) error {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.sdb == nil {
		return errors.Errorf("db not open")
	}
	path = dstore.Path(path)
	if path == "/" {
		return errors.Errorf("invalid path %s", path)
	}
	return d.set(ctx, path, values, opt...)
}

func (d *DB) set(ctx context.Context, path string, values map[string]interface{}, opt ...dstore.SetOption) error {
	opts := dstore.NewSetOptions(opt...)
	doc, err := d.get(ctx, path)
	if err != nil {
		return err
	}
	now := d.clock.Now()
	if doc == nil {
		doc = &document{
			Path:      path,
			Values:    values,
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else {
		if opts.MergeAll {
			for k, v := range values {
				doc.Values[k] = v
			}
		} else {
			doc.Values = values
		}
		doc.UpdatedAt = now
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
func (d *DB) Get(ctx context.Context, path string) (*dstore.Document, error) {
	path = dstore.Path(path)
	doc, err := d.get(ctx, path)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	return newDocument(doc), nil
}

// Load path into value.
func (d *DB) Load(ctx context.Context, path string, v interface{}) (bool, error) {
	return dstore.Load(ctx, d, path, v)
}

// GetAll paths.
func (d *DB) GetAll(ctx context.Context, paths []string) ([]*dstore.Document, error) {
	out := make([]*dstore.Document, 0, len(paths))
	for _, p := range paths {
		// TODO: Handle context Done()
		doc, err := d.get(ctx, p)
		if err != nil {
			return nil, err
		}
		if doc == nil {
			continue
		}
		out = append(out, newDocument(doc))
	}
	return out, nil
}

// Collections ...
func (d *DB) Collections(ctx context.Context, parent string) ([]*dstore.Collection, error) {
	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}
	if dstore.Path(parent) != "/" {
		// TODO: Support nested collections
		return nil, errors.Errorf("only root collections supported")
	}

	// We iterate over all the paths to build the collections list, this is slow.
	collections := []*dstore.Collection{}
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
		col := dstore.PathFirst(doc.Path)
		colv, ok := count[col]
		if !ok {
			collections = append(collections, &dstore.Collection{Path: dstore.Path(col)})
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
	path = dstore.Path(path)
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

func newDocument(doc *document) *dstore.Document {
	out := dstore.NewDocument(doc.Path)
	out.SetAll(doc.Values)
	out.CreatedAt = doc.CreatedAt
	out.UpdatedAt = doc.UpdatedAt
	return out
}

func (d *DB) unmarshal(path string, b []byte) (*document, error) {
	var doc document
	if err := msgpack.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	// Check the path requested with decrypted document path (associated data).
	if doc.Path != path {
		return nil, errors.Errorf("document path mismatch %s != %s", doc.Path, path)
	}
	return &doc, nil
}

// DocumentIterator ...
func (d *DB) DocumentIterator(ctx context.Context, parent string, opt ...dstore.Option) (dstore.Iterator, error) {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()

	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}

	opts := dstore.NewOptions(opt...)

	iter, err := d.iterator(ctx, parent, opts.Prefix)
	if err != nil {
		return nil, err
	}

	// TODO: Handle context Done()
	return &docsIterator{
		db:     d,
		iter:   iter,
		index:  opts.Index,
		limit:  opts.Limit,
		noData: opts.NoData,
	}, nil
}

func (d *DB) iterator(ctx context.Context, parent string, prefix string) (iterator.Iterator, error) {
	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}

	var iterPrefix string
	if parent != "" {
		if prefix != "" {
			iterPrefix = dstore.Path(parent, prefix)
		} else {
			iterPrefix = dstore.Path(parent) + "/"
		}
	} else {
		iterPrefix = prefix
	}

	// TODO: Handle context Done()
	return d.sdb.NewIterator(iterPrefix), nil
}

// Documents ...
func (d *DB) Documents(ctx context.Context, parent string, opt ...dstore.Option) ([]*dstore.Document, error) {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()
	opts := dstore.NewOptions(opt...)

	if opts.Index != 0 {
		return nil, errors.Errorf("index not implemented")
	}

	iter, err := d.iterator(ctx, parent, opts.Prefix)
	if err != nil {
		return nil, err
	}

	docs := []*dstore.Document{}
	for iter.Next() {
		if opts.Limit > 0 && len(docs) >= opts.Limit {
			break
		}
		path := string(iter.Key())
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		doc, err := d.unmarshal(path, iter.Value())
		if err != nil {
			return nil, err
		}
		docs = append(docs, newDocument(doc))
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return docs, nil
}

func (d *DB) get(ctx context.Context, path string) (*document, error) {
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

	return d.unmarshal(path, b)
}

// Last returns last item with key prefix.
func (d *DB) Last(ctx context.Context, prefix string) (*dstore.Document, error) {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()
	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}
	var doc *dstore.Document
	iter := d.sdb.NewIterator(prefix)
	defer iter.Release()
	if ok := iter.Last(); ok {
		path := string(iter.Key())
		val, err := d.get(ctx, path)
		if err != nil {
			return nil, err
		}
		if val != nil {
			doc = newDocument(val)
		}
	}
	if err := iter.Error(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate db")
	}
	return doc, nil
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
