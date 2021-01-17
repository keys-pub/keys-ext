package sqlcipher

import (
	"context"
	"database/sql"
	"sync"

	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
	// For sqlite3 driver
)

// ErrNotOpen if not open.
var ErrNotOpen = errors.New("not open")

// ErrAlreadyOpen if already open.
var ErrAlreadyOpen = errors.New("already open")

var _ dstore.Documents = &DB{}

// SecretKey for database.
type SecretKey *[32]byte

// DB is sqlcipher implementation of dstore.Documents.
type DB struct {
	rwmtx *sync.RWMutex
	db    *sql.DB
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
	if d.db != nil {
		return ErrAlreadyOpen
	}

	logger.Infof("Open %s", path)
	d.path = path

	db, err := open(path, key)
	if err != nil {
		return err
	}

	d.db = db
	return nil
}

// IsOpen returns true if open.
func (d *DB) IsOpen() bool {
	return d.db != nil
}

// Close the db.
func (d *DB) Close() {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()

	if d.db != nil {
		logger.Infof("Closing db %s", d.path)
		if err := d.db.Close(); err != nil {
			logger.Errorf("Error closing DB: %s", err)
		}
		d.db = nil
	}
}

// Exists returns true if the db row exists at path
func (d *DB) Exists(ctx context.Context, path string) (bool, error) {
	if d.db == nil {
		return false, ErrNotOpen
	}
	path = dstore.Path(path)
	return d.exists(path)
}

// Create entry.
func (d *DB) Create(ctx context.Context, path string, values map[string]interface{}) error {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.db == nil {
		return ErrNotOpen
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
	r := &record{
		Values:    values,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := d.insertOrReplace(path, r); err != nil {
		return err
	}

	return nil
}

// Set saves document to the db at key.
func (d *DB) Set(ctx context.Context, path string, values map[string]interface{}, opt ...dstore.SetOption) error {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	if d.db == nil {
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
	r, err := d.get(path)
	if err != nil {
		return err
	}
	now := d.clock.Now()
	if r == nil {
		r = &record{
			Values:    values,
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else {
		if opts.MergeAll {
			for k, v := range values {
				r.Values[k] = v
			}
		} else {
			r.Values = values
		}
		r.UpdatedAt = now
	}

	if err := d.insertOrReplace(path, r); err != nil {
		return err
	}

	return nil
}

// Get entry at path.
func (d *DB) Get(ctx context.Context, path string) (*dstore.Document, error) {
	path = dstore.Path(path)
	r, err := d.get(path)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	out := &dstore.Document{
		Path:      path,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	out.SetAll(r.Values)
	return out, nil
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
		r, err := d.get(p)
		if err != nil {
			return nil, err
		}
		if r == nil {
			continue
		}
		out = append(out, r.Document())
	}
	return out, nil
}

// DocumentIterator ...
func (d *DB) DocumentIterator(ctx context.Context, parent string, opt ...dstore.Option) (dstore.Iterator, error) {
	d.rwmtx.RLock()
	defer d.rwmtx.RUnlock()

	if d.db == nil {
		return nil, errors.Errorf("db not open")
	}
	parent = dstore.Path(parent)

	return d.iterator(ctx, parent, opt...)
}

// Documents ...
func (d *DB) Documents(ctx context.Context, parent string, opt ...dstore.Option) ([]*dstore.Document, error) {
	iter, err := d.DocumentIterator(ctx, parent, opt...)
	if err != nil {
		return nil, err
	}

	docs := []*dstore.Document{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if doc == nil {
			break
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// Collections ...
func (d *DB) Collections(ctx context.Context, parent string) ([]*dstore.Collection, error) {
	if d.db == nil {
		return nil, errors.Errorf("db not open")
	}
	parent = dstore.Path(parent)
	if parent != "/" {
		// TODO: Support nested collections
		return nil, errors.Errorf("only root collections supported")
	}

	// We iterate over all the paths to build the collections list, this is slow.
	collections := []*dstore.Collection{}
	count := map[string]int{}
	iter, err := d.iterator(ctx, parent)
	if err != nil {
		return nil, err
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
	if d.db == nil {
		return false, errors.Errorf("db not open")
	}
	path = dstore.Path(path)
	logger.Infof("Deleting %s", path)
	return d.delete(path)
}

// DeleteAll paths.
func (d *DB) DeleteAll(ctx context.Context, paths []string) error {
	for _, p := range paths {
		if _, err := d.Delete(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

// DeleteCollection to at paths at parent.
func (d *DB) DeleteCollection(ctx context.Context, parent string) error {
	docs, err := d.Documents(ctx, parent, dstore.NoData())
	if err != nil {
		return err
	}
	paths := []string{}
	for _, d := range docs {
		paths = append(paths, d.Path)
	}
	return d.DeleteAll(ctx, paths)
}
