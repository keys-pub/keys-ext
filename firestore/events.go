package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

// eventIdxLabel should also match the Event firestore tag.
const eventIdxLabel = "idx"

func (f *Firestore) EventAdd(ctx context.Context, path string, doc events.Document) (int64, error) {
	idx, err := f.EventsAdd(ctx, path, []events.Document{doc})
	return idx, err
}

func (f *Firestore) EventsAdd(ctx context.Context, path string, docs []events.Document) (int64, error) {
	pos := 0
	remaining := len(docs)
	idx := int64(0)
	for remaining > 0 {
		chunk := min(498, remaining)
		logger.Infof(ctx, "Writing %s (batch %d:%d)", path, pos, pos+chunk)
		bidx, err := f.writeAll(ctx, normalizePath(path), docs[pos:pos+chunk])
		if err != nil {
			// TODO: Delete previous batch writes if pos > 0
			return 0, errors.Wrapf(err, "failed to write")
		}
		pos = pos + chunk
		remaining = remaining - chunk
		idx = bidx
	}
	return idx, nil
}

func (f *Firestore) EventPosition(ctx context.Context, path string) (*events.Position, error) {
	res, err := f.EventPositions(ctx, []string{path})
	if err != nil {
		return nil, err
	}
	return res[path], nil
}

// EventPositions returns positions for event logs.
func (f *Firestore) EventPositions(ctx context.Context, paths []string) (map[string]*events.Position, error) {
	positions := map[string]*events.Position{}
	docs, err := f.GetAll(ctx, paths)
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		idx, _ := doc.Int64("idx")
		positions[doc.Path] = &events.Position{
			Path:      doc.Path,
			Index:     int64(idx),
			Timestamp: tsutil.Millis(doc.UpdatedAt),
		}
	}
	return positions, nil
}

func (f *Firestore) writeAll(ctx context.Context, path string, docs []events.Document) (int64, error) {
	if len(docs) > 498 {
		return 0, errors.Errorf("too many events (max 498)")
	}

	index := int64(0)
	if err := f.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Get current event index
		res, err := f.txGet(tx, path)
		if err != nil {
			return err
		}
		if res != nil {
			i, ok := res.Data()[eventIdxLabel]
			if ok {
				index = i.(int64)
			}
		} else {
			if err := tx.Create(f.client.Doc(path), map[string]interface{}{
				eventIdxLabel: 0,
			}); err != nil {
				return err
			}
		}

		// Write docs (incrementing index)
		for _, doc := range docs {
			id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
			lpath := dstore.Path(path, "log", id)

			index++
			docRef := f.client.Doc(normalizePath(lpath))
			doc[eventIdxLabel] = index
			if err := tx.Create(docRef, doc); err != nil {
				return err
			}
		}

		// Update index
		if err := tx.Set(f.client.Doc(path), map[string]interface{}{
			eventIdxLabel: index,
		}, firestore.MergeAll); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return 0, err
	}
	return index, nil
}

// Events ...
func (f *Firestore) Events(ctx context.Context, path string, opt ...events.Option) (events.Iterator, error) {
	opts := events.NewOptions(opt...)
	log := normalizePath(dstore.Path(path, "log"))
	col := f.client.Collection(log)
	if col == nil {
		return events.NewIterator([]*events.Event{}), nil
	}

	var q firestore.Query
	switch opts.Direction {
	case events.Ascending:
		if opts.Index == 0 {
			logger.Infof(ctx, "List events %s (asc)...", log)
			q = col.OrderBy(eventIdxLabel, firestore.Asc)
		} else {
			logger.Infof(ctx, "List events %s (asc > %d)", log, opts.Index)
			q = col.OrderBy(eventIdxLabel, firestore.Asc).Where(eventIdxLabel, ">", opts.Index)
		}
	case events.Descending:
		if opts.Index == 0 {
			logger.Infof(ctx, "List events %s (desc)...", log)
			q = col.OrderBy(eventIdxLabel, firestore.Desc)
		} else {
			logger.Infof(ctx, "List events %s (desc < %d)", log, opts.Index)
			q = col.OrderBy(eventIdxLabel, firestore.Desc).Where(eventIdxLabel, "<", opts.Index)
		}
	}

	iter := q.Documents(ctx)

	return &eventIterator{iter: iter, limit: opts.Limit}, nil
}

// EventsDelete removes events.
func (f *Firestore) EventsDelete(ctx context.Context, path string) (bool, error) {
	log := dstore.Path(path, "log")
	if err := f.deleteCollection(ctx, log, 100); err != nil {
		return false, err
	}

	doc := f.client.Doc(normalizePath(path))
	exists, err := f.Exists(ctx, path)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	if _, err := doc.Delete(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (f *Firestore) deleteCollection(ctx context.Context, path string, batchSize int) error {
	col := f.client.Collection(normalizePath(path))

	// From https://firebase.google.com/docs/firestore/manage-data/delete-data#go_2
	for {
		// Get a batch of documents
		iter := col.Limit(batchSize).Documents(ctx)
		numDeleted := 0

		// Iterate through the documents, adding
		// a delete operation for each one to a
		// WriteBatch.
		batch := f.client.Batch()
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}

			batch.Delete(doc.Ref)
			numDeleted++
		}

		// If there are no documents to delete,
		// the process is over.
		if numDeleted == 0 {
			return nil
		}

		_, err := batch.Commit(ctx)
		if err != nil {
			return err
		}
	}
}

func (f *Firestore) Increment(ctx context.Context, path string, name string, n int64) (int64, int64, error) {
	if n < 1 {
		return 0, 0, errors.Errorf("increment by at least 1")
	}
	path = normalizePath(path)
	index := int64(0)
	if err := f.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Get current index
		res, err := f.txGet(tx, path)
		if err != nil {
			return err
		}
		if res != nil {
			i, ok := res.Data()[name]
			if ok {
				index = i.(int64)
			}
		} else {
			if err := tx.Create(f.client.Doc(path), map[string]interface{}{
				name: 0,
			}); err != nil {
				return err
			}
		}
		index += n

		// Update index
		if err := tx.Set(f.client.Doc(path), map[string]interface{}{
			name: index,
		}, firestore.MergeAll); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, 0, err
	}
	return index, index - n + 1, nil
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
