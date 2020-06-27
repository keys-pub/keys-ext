package firestore

import (
	"context"
	"fmt"
	"testing"

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
		require.False(t, event.Timestamp.IsZero())
		require.Equal(t, int64(i+1), event.Index)
	}

	// Events (limit=10, asc)
	iter, err := eds.Events(ctx, path, 0, 10, ds.Ascending)
	require.NoError(t, err)
	eventsValues := []string{}
	index := int64(0)
	for i := 0; ; i++ {
		event, err := iter.Next()
		require.NoError(t, err)
		if event == nil {
			break
		}
		require.False(t, event.Timestamp.IsZero())
		require.Equal(t, int64(i+1), event.Index)
		eventsValues = append(eventsValues, string(event.Data))
		index = event.Index
	}
	iter.Release()
	require.Equal(t, 10, len(eventsValues))
	require.Equal(t, strs[0:10], eventsValues)

	// Events (index, asc)
	iter, err = eds.Events(ctx, path, index, 10, ds.Ascending)
	require.NoError(t, err)
	eventsValues = []string{}
	for i := 0; ; i++ {
		event, err := iter.Next()
		require.NoError(t, err)
		if event == nil {
			break
		}
		eventsValues = append(eventsValues, string(event.Data))
		index = event.Index
	}
	iter.Release()
	require.Equal(t, int64(20), index)
	require.Equal(t, 10, len(eventsValues))

	require.Equal(t, strs[10:20], eventsValues)

	// Events (large index)
	large := int64(1000000000)
	iter, err = eds.Events(ctx, path, large, 100, ds.Ascending)
	require.NoError(t, err)
	event, err := iter.Next()
	require.NoError(t, err)
	require.Nil(t, event)
	iter.Release()

	// Descending
	revs := reverseCopy(strs)

	// Events (limit=10, desc)
	iter, err = eds.Events(ctx, path, 0, 10, ds.Descending)
	require.NoError(t, err)
	eventsValues = []string{}
	for i := 0; ; i++ {
		event, err := iter.Next()
		require.NoError(t, err)
		if event == nil {
			break
		}
		eventsValues = append(eventsValues, string(event.Data))
		index = event.Index
	}
	iter.Release()
	require.Equal(t, 10, len(eventsValues))
	require.Equal(t, int64(31), index)
	require.Equal(t, revs[0:10], eventsValues)

	// Events (limit=5, index, desc)
	iter, err = eds.Events(ctx, path, index, 5, ds.Descending)
	require.NoError(t, err)
	eventsValues = []string{}
	for i := 0; ; i++ {
		event, err := iter.Next()
		require.NoError(t, err)
		if event == nil {
			break
		}
		eventsValues = append(eventsValues, string(event.Data))
		index = event.Index
	}
	iter.Release()
	require.Equal(t, 5, len(eventsValues))
	require.Equal(t, int64(26), index)
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
