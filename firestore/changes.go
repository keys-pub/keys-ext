package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

// ChangesAdd adds changes.
func (f *Firestore) ChangesAdd(ctx context.Context, collection string, data [][]byte) error {
	if f.incrementFn == nil {
		return errors.Errorf("no increment fn set")
	}
	if len(data) > 499 {
		return errors.Errorf("too many changes to batch (max 500)")
	}

	batch := f.client.Batch()

	for _, b := range data {
		version, err := f.incrementFn(ctx)
		if err != nil {
			return err
		}

		id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
		path := ds.Path(collection, id)
		// Map should match keys.
		m := map[string]interface{}{
			"data": b,
			"v":    version,
			"ts":   firestore.ServerTimestamp,
		}
		doc := f.client.Doc(normalizePath(path))
		batch.Create(doc, m)
	}

	if _, err := batch.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// Changes ...
func (f *Firestore) Changes(ctx context.Context, collection string, version int64, limit int, direction ds.Direction) (ds.ChangeIterator, error) {
	path := normalizePath(collection)
	col := f.client.Collection(path)
	if col == nil {
		return ds.NewChangeIterator([]*ds.Change{}), nil
	}

	var q firestore.Query
	switch direction {
	case ds.Ascending:
		if version == 0 {
			logger.Infof(ctx, "List changes (asc)...")
			q = col.OrderBy("v", firestore.Asc)
		} else {
			logger.Infof(ctx, "List changes (asc > %s)", version)
			q = col.OrderBy("v", firestore.Asc).Where("v", ">", version)
		}
	case ds.Descending:
		if version == 0 {
			logger.Infof(ctx, "List changes (desc)...")
			q = col.OrderBy("v", firestore.Desc)
		} else {
			logger.Infof(ctx, "List changes (desc < %s)", version)
			q = col.OrderBy("v", firestore.Desc).Where("v", "<", version)
		}
	}

	iter := q.Documents(ctx)

	// TODO: Put limits when clients can handle paging
	// if limit == 0 {
	// 	limit = 100
	// }

	return &changeIterator{iter: iter, limit: limit}, nil
}
