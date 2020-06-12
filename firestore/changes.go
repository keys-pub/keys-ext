package firestore

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/encoding"
)

// timestampField should match firestore tag on keys.Change.
const timestampField = "ts"

// ChangeAdd adds Change.
func (f *Firestore) ChangeAdd(ctx context.Context, collection string, data []byte) (string, error) {
	id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	path := ds.Path(collection, id)
	// Map should match keys.Change json format
	m := map[string]interface{}{
		"data":         data,
		timestampField: firestore.ServerTimestamp,
	}
	if err := f.createValue(ctx, path, m); err != nil {
		return "", err
	}
	return path, nil
}

// Changes ...
func (f *Firestore) Changes(ctx context.Context, collection string, ts time.Time, limit int, direction ds.Direction) (ds.ChangeIterator, error) {
	path := normalizePath(collection)
	col := f.client.Collection(path)
	if col == nil {
		return ds.NewChangeIterator([]*ds.Change{}), nil
	}

	var q firestore.Query
	switch direction {
	case ds.Ascending:
		if ts.IsZero() {
			logger.Infof(ctx, "List changes (asc)...")
			q = col.OrderBy(timestampField, firestore.Asc)
		} else {
			logger.Infof(ctx, "List changes (asc >= %s)", ts)
			q = col.OrderBy(timestampField, firestore.Asc).Where(timestampField, ">=", ts)
		}
	case ds.Descending:
		if ts.IsZero() {
			logger.Infof(ctx, "List changes (desc)...")
			q = col.OrderBy(timestampField, firestore.Desc)
		} else {
			logger.Infof(ctx, "List changes (desc <= %s)", ts)
			q = col.OrderBy(timestampField, firestore.Desc).Where(timestampField, "<=", ts)
		}
	}

	iter := q.Documents(ctx)

	if limit == 0 {
		limit = 100
	}

	return &changeIterator{iter: iter, limit: limit}, nil
}
