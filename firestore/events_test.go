package firestore

import (
	"context"
	"fmt"
	"testing"

	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/stretchr/testify/require"
)

func TestEvents(t *testing.T) {
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
	out, idx, err := eds.EventsAdd(ctx, path, values)
	require.NoError(t, err)
	require.Equal(t, 40, len(out))
	require.Equal(t, int64(40), idx)
	for i, event := range out {
		require.NotEmpty(t, event.Timestamp)
		require.Equal(t, int64(i+1), event.Index)
	}

	// Events (limit=10, asc)
	iter, err := eds.Events(ctx, path, events.Limit(10))
	require.NoError(t, err)
	eventsValues := []string{}
	index := int64(0)
	for i := 0; ; i++ {
		event, err := iter.Next()
		require.NoError(t, err)
		if event == nil {
			break
		}
		require.NotEmpty(t, event.Timestamp)
		require.Equal(t, int64(i+1), event.Index)
		eventsValues = append(eventsValues, string(event.Data))
		index = event.Index
	}
	iter.Release()
	require.Equal(t, 10, len(eventsValues))
	require.Equal(t, strs[0:10], eventsValues)

	// Events (index, asc)
	iter, err = eds.Events(ctx, path, events.Index(index), events.Limit(10))
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
	iter, err = eds.Events(ctx, path, events.Index(large))
	require.NoError(t, err)
	event, err := iter.Next()
	require.NoError(t, err)
	require.Nil(t, event)
	iter.Release()

	// Descending
	revs := reverseCopy(strs)

	// Events (limit=10, desc)
	iter, err = eds.Events(ctx, path, events.Limit(10), events.WithDirection(events.Descending))
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
	iter, err = eds.Events(ctx, path, events.Index(index), events.Limit(5), events.WithDirection(events.Descending))
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

	positions, err := eds.EventPositions(ctx, []string{path})
	require.NoError(t, err)
	require.Equal(t, 1, len(positions))
	require.Equal(t, int64(40), positions[path].Index)

	// Delete
	ok, err := eds.EventsDelete(ctx, path)
	require.NoError(t, err)
	require.True(t, ok)

	iter, err = eds.Events(ctx, path)
	require.NoError(t, err)
	event, err = iter.Next()
	require.NoError(t, err)
	require.Nil(t, event)
	iter.Release()

	positions, err = eds.EventPositions(ctx, []string{path})
	require.NoError(t, err)
	require.Equal(t, 0, len(positions))

	ok, err = eds.EventsDelete(ctx, path)
	require.NoError(t, err)
	require.False(t, ok)
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

func TestFirestoreBatch(t *testing.T) {
	var err error
	// SetContextLogger(NewContextLogger(DebugLevel))

	eds := testFirestore(t)
	ctx := context.TODO()
	path := testPath()
	t.Logf("Path: %s", path)

	values := [][]byte{}
	length := 1001
	for i := 0; i < length; i++ {
		str := fmt.Sprintf("value%d", i)
		values = append(values, []byte(str))
	}
	out, idx, err := eds.EventsAdd(ctx, path, values)
	require.NoError(t, err)
	require.Equal(t, length, len(out))
	require.Equal(t, int64(length), idx)

	iter, err := eds.Events(ctx, path)
	require.NoError(t, err)
	i := 0
	for {
		event, err := iter.Next()
		require.NoError(t, err)
		if event == nil {
			break
		}
		require.Equal(t, fmt.Sprintf("value%d", i), string(event.Data))
		i++
	}
	iter.Release()
	require.Equal(t, length, i)
}

func TestUpdateWithEvents(t *testing.T) {
	ds := testFirestore(t)
	ctx := context.TODO()
	collection := testCollection()

	path := dstore.Path(collection, "key1")

	_, _, err := ds.EventsAdd(ctx, path, [][]byte{[]byte("test1"), []byte("test2")})
	require.NoError(t, err)

	err = ds.Set(ctx, path, map[string]interface{}{"info": "testinfo", "data": []byte("val1")}, dstore.MergeAll())
	require.NoError(t, err)

	doc, err := ds.Get(ctx, path)
	require.NoError(t, err)
	require.NotNil(t, doc)

	b := doc.Bytes("data")
	require.Equal(t, []byte("val1"), b)

	index, _ := doc.Int("idx") // From events
	require.Equal(t, 2, index)

	info, _ := doc.String("info")
	require.Equal(t, "testinfo", info)

	// Events
	iter, err := ds.Events(ctx, path, events.Limit(10))
	require.NoError(t, err)
	event1, err := iter.Next()
	require.NoError(t, err)
	require.Equal(t, []byte("test1"), event1.Data)
	event2, err := iter.Next()
	require.NoError(t, err)
	require.Equal(t, []byte("test2"), event2.Data)
}
