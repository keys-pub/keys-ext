package firestore

import (
	"context"
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys/ds"
	"github.com/stretchr/testify/require"
)

func TestFirestoreEvents(t *testing.T) {
	var err error
	// SetContextLogger(NewContextLogger(DebugLevel))
	eds := testFirestore(t)
	ctx := context.TODO()
	path := testPath()
	t.Logf("Path: %s", path)

	length := 40
	values := [][]byte{}
	strs := []string{}
	for i := 0; i < length; i++ {
		str := fmt.Sprintf("value%d", i)
		values = append(values, []byte(str))
		strs = append(strs, str)
	}
	out, err := eds.EventsAdd(ctx, path, values)
	require.NoError(t, err)
	require.Equal(t, 40, len(out))
	for i, event := range out {
		t.Logf("Event: %s", spew.Sdump(event))
		require.False(t, event.Timestamp.IsZero())
		require.Equal(t, int64(i+1), event.Index)
	}

	// Events (limit=10, asc)
	iter, err := eds.Events(ctx, path, 0, 10, ds.Ascending)
	require.NoError(t, err)
	events, index, err := ds.EventsFromIterator(iter, 0)
	require.NoError(t, err)
	iter.Release()
	require.Equal(t, 10, len(events))
	eventsValues := []string{}
	for i, event := range events {
		require.False(t, event.Timestamp.IsZero())
		require.Equal(t, int64(i+1), event.Index)
		eventsValues = append(eventsValues, string(event.Data))
	}
	require.Equal(t, strs[0:10], eventsValues)

	// Events (index, asc)
	iter, err = eds.Events(ctx, path, index, 10, ds.Ascending)
	require.NoError(t, err)
	events, index, err = ds.EventsFromIterator(iter, index)
	require.NoError(t, err)
	iter.Release()
	require.Equal(t, int64(20), index)
	require.Equal(t, 10, len(events))
	eventsValues = []string{}
	for _, event := range events {
		eventsValues = append(eventsValues, string(event.Data))
	}
	require.Equal(t, strs[10:20], eventsValues)

	// Events (large index)
	large := int64(1000000000)
	iter, err = eds.Events(ctx, path, large, 100, ds.Ascending)
	require.NoError(t, err)
	events, index, err = ds.EventsFromIterator(iter, large)
	require.NoError(t, err)
	iter.Release()
	require.Equal(t, 0, len(events))
	require.Equal(t, large, index)

	// Descending
	revs := reverseCopy(strs)

	// Events (limit=10, desc)
	iter, err = eds.Events(ctx, path, 0, 10, ds.Descending)
	require.NoError(t, err)
	events, index, err = ds.EventsFromIterator(iter, 0)
	require.NoError(t, err)
	iter.Release()
	require.Equal(t, 10, len(events))
	require.Equal(t, int64(31), index)
	eventsValues = []string{}
	for _, event := range events {
		eventsValues = append(eventsValues, string(event.Data))
	}
	require.Equal(t, revs[0:10], eventsValues)

	// Events (limit=5, index, desc)
	iter, err = eds.Events(ctx, path, index, 5, ds.Descending)
	require.NoError(t, err)
	events, index, err = ds.EventsFromIterator(iter, index)
	require.NoError(t, err)
	iter.Release()
	require.Equal(t, 5, len(events))
	require.Equal(t, int64(26), index)
	eventsValues = []string{}
	for _, event := range events {
		eventsValues = append(eventsValues, string(event.Data))
	}
	require.Equal(t, revs[10:15], eventsValues)
}

func TestIndex(t *testing.T) {
	var err error
	// SetContextLogger(NewContextLogger(DebugLevel))
	eds := testFirestore(t)
	ctx := context.TODO()
	path := testPath()

	ver, err := eds.index(ctx, path, 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), ver)

	ver, err = eds.index(ctx, path, 5)
	require.NoError(t, err)
	require.Equal(t, int64(2), ver)

	ver, err = eds.index(ctx, path, 3)
	require.NoError(t, err)
	require.Equal(t, int64(7), ver)

	ver, err = eds.index(ctx, path, 1)
	require.NoError(t, err)
	require.Equal(t, int64(10), ver)
}

func reverseCopy(s []string) []string {
	a := make([]string, len(s))
	for i, j := 0, len(s)-1; i < len(s); i++ {
		a[i] = s[j]
		j--
	}
	return a
}
