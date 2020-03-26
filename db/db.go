package db

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var _ keys.DocumentStore = &DB{}

// DB is leveldb implementation of keys.DocumentStore.
type DB struct {
	rwmtx *sync.RWMutex
	sdb   *sdb
	fpath string
	nowFn func() time.Time

	key keys.SecretKey
}

// NewDB creates a DB.
func NewDB() *DB {
	return &DB{
		rwmtx: &sync.RWMutex{},
		nowFn: time.Now,
	}
}

// SetTimeNow sets clock.
func (d *DB) SetTimeNow(nowFn func() time.Time) {
	d.nowFn = nowFn
}

// Now returns current time.
func (d *DB) Now() time.Time {
	return d.nowFn()
}

// OpenAtPath opens db located at path
func (d *DB) OpenAtPath(ctx context.Context, path string, key keys.SecretKey) error {
	logger.Infof("LevelDB at %s", path)
	d.fpath = path
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return err
	}
	sdb := newSDB(db, key)
	d.sdb = sdb
	d.key = key
	return nil
}

// Close the db.
func (d *DB) Close() {
	logger.Infof("Closing leveldb %s", d.fpath)
	if err := d.sdb.Close(); err != nil {
		logger.Errorf("Error closing DB: %s", err)
	}
}

// Exists returns true if the db row exists at path
func (d *DB) Exists(ctx context.Context, path string) (bool, error) {
	if d.sdb == nil {
		return false, errors.Errorf("db not open")
	}
	path = keys.Path(path)
	return d.sdb.Has(path)
}

// Create entry.
func (d *DB) Create(ctx context.Context, path string, b []byte) error {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.sdb == nil {
		return errors.Errorf("db not open")
	}
	path = keys.Path(path)
	if path == "/" {
		return errors.Errorf("invalid path %s", path)
	}
	exists, err := d.Exists(ctx, path)
	if err != nil {
		return err
	}
	if exists {
		return keys.NewErrPathExists(path)
	}

	now := d.Now()
	md := &metadata{CreateTime: now, UpdateTime: now}
	return d.put(path, b, md)
}

// Set saves document to the db at key.
func (d *DB) Set(ctx context.Context, path string, b []byte) error {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.sdb == nil {
		return errors.Errorf("db not open")
	}
	path = keys.Path(path)
	if path == "/" {
		return errors.Errorf("invalid path %s", path)
	}

	md, err := d.getMetadata(path)
	if err != nil {
		return err
	}
	if md == nil {
		md = &metadata{}
	}
	now := d.Now()
	if md.CreateTime.IsZero() {
		md.CreateTime = now
	}
	md.UpdateTime = now

	return d.put(path, b, md)
}

func (d *DB) put(path string, b []byte, md *metadata) error {
	if err := d.setMetadata(path, md); err != nil {
		return err
	}
	if err := d.setCollection(path, md); err != nil {
		return err
	}
	logger.Debugf("Put %s (%d bytes)", path, len(b))
	if err := d.sdb.Put(path, b); err != nil {
		return err
	}
	return nil
}

type metadata struct {
	CreateTime time.Time
	UpdateTime time.Time
}

func (d *DB) getMetadata(path string) (*metadata, error) {
	b, err := d.sdb.Get("~" + path)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	var md metadata
	if err := json.Unmarshal(b, &md); err != nil {
		return nil, err
	}
	return &md, nil
}

func (d *DB) setMetadata(path string, md *metadata) error {
	mpath := "~" + path
	logger.Debugf("Set metadata %s %+v", mpath, md)
	b, err := json.Marshal(md)
	if err != nil {
		return err
	}
	return d.sdb.Put(mpath, b)
}

func (d *DB) setCollection(path string, md *metadata) error {
	cpath := "+" + keys.FirstPathComponent(path)
	logger.Debugf("Set collection %s %+v", cpath, md)
	b, err := json.Marshal(md)
	if err != nil {
		return err
	}
	return d.sdb.Put(cpath, b)
}

// Get entry at path.
func (d *DB) Get(ctx context.Context, path string) (*keys.Document, error) {
	path = keys.Path(path)
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
func (d *DB) GetAll(ctx context.Context, paths []string) ([]*keys.Document, error) {
	docs := make([]*keys.Document, 0, len(paths))
	for _, p := range paths {
		// TODO: Handle context Done()
		doc, err := d.get(ctx, p)
		if err != nil {
			return nil, err
		}
		if doc == nil {
			continue
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// Collections ...
func (d *DB) Collections(ctx context.Context, parent string) (keys.CollectionIterator, error) {
	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}
	if keys.Path(parent) != "/" {
		return nil, errors.Errorf("only root collections supported")
	}

	iter := d.sdb.NewIterator("+")
	return &colsIterator{iter: iter}, nil
}

// Delete value at path.
func (d *DB) Delete(ctx context.Context, path string) (bool, error) {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.sdb == nil {
		return false, errors.Errorf("db not open")
	}
	path = keys.Path(path)
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

// DeleteAll deletes values with key prefix.
func (d *DB) DeleteAll(ctx context.Context, parent string) error {
	iter, err := d.Documents(ctx, parent, &keys.DocumentsOpts{PathOnly: true})
	if err != nil {
		return err
	}
	for {
		doc, err := iter.Next()
		if err != nil {
			return err
		}
		if doc == nil {
			break
		}
		if _, err := d.Delete(ctx, doc.Path); err != nil {
			return err
		}
	}
	iter.Release()
	return nil
}

func (d *DB) document(path string, b []byte) (*keys.Document, error) {
	md, err := d.getMetadata(path)
	if err != nil {
		return nil, err
	}
	doc := keys.NewDocument(path, b)
	doc.CreatedAt = md.CreateTime
	doc.UpdatedAt = md.UpdateTime
	return doc, nil
}

// Documents ...
func (d *DB) Documents(ctx context.Context, parent string, opts *keys.DocumentsOpts) (keys.DocumentIterator, error) {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()
	if opts == nil {
		opts = &keys.DocumentsOpts{}
	}

	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}

	path := keys.Path(parent)

	var prefix string
	if opts.Prefix != "" {
		prefix = keys.Path(path, opts.Prefix)
	} else if path != "/" {
		prefix = path + "/"
	} else {
		prefix = path
	}

	if path == "/" {
		return nil, errors.Errorf("list root not supported")
	}

	logger.Debugf("Iterator prefix %s", prefix)
	// TODO: Handle context Done()
	iter := d.sdb.NewIterator(prefix)
	return &docsIterator{
		db:    d,
		iter:  iter,
		index: opts.Index,
		limit: opts.Limit,
	}, nil
}

func (d *DB) get(ctx context.Context, path string) (*keys.Document, error) {
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
func (d *DB) Last(ctx context.Context, prefix string) (*keys.Document, error) {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()
	if d.sdb == nil {
		return nil, errors.Errorf("db not open")
	}
	var doc *keys.Document
	iter := d.sdb.NewIterator(prefix)
	if ok := iter.Last(); ok {
		path := string(iter.Value())
		val, err := d.get(ctx, path)
		if err != nil {
			return nil, err
		}
		doc = val
	}
	iter.Release()
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
	total := 0
	for iter.Next() {
		path := string(iter.Key())
		if contains != "" {
			value := iter.Value()
			entry := keys.NewDocument(path, value)
			if entry.Contains(contains) {
				total++
			}
		} else {
			total++
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return -1, err
	}
	return total, nil
}
