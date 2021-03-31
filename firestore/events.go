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

// EventsAdd adds events.
func (f *Firestore) EventsAdd(ctx context.Context, path string, data [][]byte) ([]*events.Event, int64, error) {
	pos := 0
	remaining := len(data)
	events := make([]*events.Event, 0, len(data))
	idx := int64(0)
	for remaining > 0 {
		chunk := min(500, remaining)
		logger.Infof(ctx, "Writing %s (batch %d:%d)", path, pos, pos+chunk)
		batch, bidx, err := f.writeBatch(ctx, path, data[pos:pos+chunk])
		if err != nil {
			// TODO: Delete previous batch writes if pos > 0
			return nil, 0, errors.Wrapf(err, "failed to write batch")
		}
		events = append(events, batch...)
		pos = pos + chunk
		remaining = remaining - chunk
		idx = bidx
	}
	return events, idx, nil
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

func (f *Firestore) writeBatch(ctx context.Context, path string, data [][]byte) ([]*events.Event, int64, error) {
	if len(data) > 500 {
		return nil, 0, errors.Errorf("too many events to batch (max 500)")
	}

	idx, err := f.Increment(ctx, dstore.Path(path), eventIdxLabel, int64(len(data)))
	if err != nil {
		return nil, 0, errors.Wrapf(err, "failed to increment index")
	}

	batch := f.client.Batch()

	out := make([]*events.Event, 0, len(data))
	last := int64(0)
	for _, b := range data {
		id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
		path := dstore.Path(path, "log", id)
		// logger.Debugf(ctx, "Batching %s (%d)", path, idx)

		// Map should match keys.
		m := map[string]interface{}{
			"data":        b,
			eventIdxLabel: idx,
		}
		doc := f.client.Doc(normalizePath(path))
		batch.Create(doc, m)

		out = append(out, &events.Event{
			Data:  b,
			Index: idx,
		})
		last = idx
		idx++
	}

	res, err := batch.Commit(ctx)
	if err != nil {
		return nil, 0, err
	}
	for i, event := range out {
		event.Timestamp = tsutil.Millis(res[i].UpdateTime)
	}

	return out, last, nil
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

// Increment is a very slow increment by. Limited to 1 write a second.
// Returns start of index.
// If we need better performance we can shard.
// TODO: https://firebase.google.com/docs/firestore/solutions/counters#go
func (f *Firestore) Increment(ctx context.Context, path string, name string, n int64) (int64, error) {
	if n < 1 {
		return 0, errors.Errorf("increment by at least 1")
	}
	exists, err := f.Exists(ctx, path)
	if err != nil {
		return 0, err
	}
	count := f.client.Doc(normalizePath(path))
	if !exists {
		if _, err := count.Create(ctx, map[string]interface{}{name: 0}); err != nil {
			return 0, err
		}
	}
	if _, err := count.Update(ctx, []firestore.Update{
		{Path: name, Value: firestore.Increment(n)},
	}); err != nil {
		return 0, err
	}
	res, err := count.Get(ctx)
	if err != nil {
		return 0, err
	}
	index := res.Data()[name].(int64)

	return index - n + 1, nil
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
