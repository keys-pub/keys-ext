package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

// eventIdx should also match ds.Event firestore tag.
const eventIdxLabel = "idx"

// EventsAdd adds events.
func (f *Firestore) EventsAdd(ctx context.Context, path string, data [][]byte) ([]*ds.Event, error) {
	if len(data) > 499 {
		return nil, errors.Errorf("too many events to batch (max 500)")
	}

	batch := f.client.Batch()

	idx, err := f.index(ctx, ds.Path(path), int64(len(data)))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to increment index")
	}

	events := make([]*ds.Event, 0, len(data))
	for _, b := range data {
		id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
		path := ds.Path(path, "log", id)
		logger.Debugf(ctx, "Batching %s (%d)", path, idx)

		// Map should match keys.
		m := map[string]interface{}{
			"data":        b,
			eventIdxLabel: idx,
		}
		doc := f.client.Doc(normalizePath(path))
		batch.Create(doc, m)

		events = append(events, &ds.Event{
			Data:  b,
			Index: idx,
		})

		idx++
	}

	res, err := batch.Commit(ctx)
	if err != nil {
		return nil, err
	}
	for i, event := range events {
		event.Timestamp = res[i].UpdateTime
	}

	return events, nil
}

// Events ...
func (f *Firestore) Events(ctx context.Context, path string, index int64, limit int, direction ds.Direction) (ds.EventIterator, error) {
	col := f.client.Collection(normalizePath(ds.Path(path, "log")))
	if col == nil {
		return ds.NewEventIterator([]*ds.Event{}), nil
	}

	var q firestore.Query
	switch direction {
	case ds.Ascending:
		if index == 0 {
			logger.Infof(ctx, "List events (asc)...")
			q = col.OrderBy(eventIdxLabel, firestore.Asc)
		} else {
			logger.Infof(ctx, "List events (asc > %d)", index)
			q = col.OrderBy(eventIdxLabel, firestore.Asc).Where(eventIdxLabel, ">", index)
		}
	case ds.Descending:
		if index == 0 {
			logger.Infof(ctx, "List events (desc)...")
			q = col.OrderBy(eventIdxLabel, firestore.Desc)
		} else {
			logger.Infof(ctx, "List events (desc < %d)", index)
			q = col.OrderBy(eventIdxLabel, firestore.Desc).Where(eventIdxLabel, "<", index)
		}
	}

	iter := q.Documents(ctx)

	// TODO: Put limits when clients can handle paging
	// if limit == 0 {
	// 	limit = 100
	// }

	return &eventIterator{iter: iter, limit: limit}, nil
}

// index is a very slow increment by. Limited to 1 write a second.
// Returns start of index.
// If we need better performance we can shard.
// TODO: https://firebase.google.com/docs/firestore/solutions/counters#go
func (f *Firestore) index(ctx context.Context, path string, n int64) (int64, error) {
	if n < 1 {
		return 0, errors.Errorf("increment by at least 1")
	}
	exists, err := f.Exists(ctx, path)
	if err != nil {
		return 0, err
	}
	count := f.client.Doc(normalizePath(path))
	if !exists {
		if _, err := count.Create(ctx, map[string]interface{}{eventIdxLabel: 0}); err != nil {
			return 0, err
		}
	}
	if _, err := count.Update(ctx, []firestore.Update{
		{Path: eventIdxLabel, Value: firestore.Increment(n)},
	}); err != nil {
		return 0, err
	}
	res, err := count.Get(ctx)
	if err != nil {
		return 0, err
	}
	index := res.Data()[eventIdxLabel].(int64)

	return index - n + 1, nil
}
