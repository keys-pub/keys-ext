package firestore

import (
	"context"
	"fmt"
	"sync"
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
	values := []events.Document{}
	strs := []string{}
	for i := 0; i < length; i++ {
		str := fmt.Sprintf("value%d", i)
		values = append(values, dstore.Data([]byte(str)))
		strs = append(strs, str)
	}
	idx, err := eds.EventsAdd(ctx, path, values)
	require.NoError(t, err)
	require.Equal(t, int64(40), idx)

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
		eventsValues = append(eventsValues, string(event.Data()))
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
		eventsValues = append(eventsValues, string(event.Data()))
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
		eventsValues = append(eventsValues, string(event.Data()))
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
		eventsValues = append(eventsValues, string(event.Data()))
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

func TestEventsConcurrent(t *testing.T) {
	eds := testFirestore(t)
	ctx := context.TODO()
	path := testPath()
	t.Logf("Path: %s", path)

	_, err := eds.EventAdd(ctx, path, dstore.Data([]byte("testing")))
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(4)

	fn := func(group string) {
		var ferr error
		for i := 1; i < 5; i++ {
			val := fmt.Sprintf("testing-%s-%d", group, i)
			_, ferr = eds.EventAdd(ctx, path, dstore.Data([]byte(val)))
			if ferr != nil {
				break
			}
		}
		wg.Done()
		require.NoError(t, ferr)
	}
	go fn("a")
	go fn("b")
	go fn("c")
	go fn("d")

	wg.Wait()

	idx := int64(1)
	iter, err := eds.Events(ctx, path)
	require.NoError(t, err)
	defer iter.Release()
	for {
		event, err := iter.Next()
		require.NoError(t, err)
		if event == nil {
			break
		}
		require.Equal(t, idx, event.Index)
		idx++
	}
}

func TestIncrementIndex(t *testing.T) {
	var err error
	// SetContextLogger(NewContextLogger(DebugLevel))
	eds := testFirestore(t)
	ctx := context.TODO()
	path := testPath()

	_, idx, err := eds.Increment(ctx, path, "idx", 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), idx)

	_, idx, err = eds.Increment(ctx, path, "idx", 5)
	require.NoError(t, err)
	require.Equal(t, int64(2), idx)

	_, idx, err = eds.Increment(ctx, path, "idx", 3)
	require.NoError(t, err)
	require.Equal(t, int64(7), idx)

	_, idx, err = eds.Increment(ctx, path, "idx", 1)
	require.NoError(t, err)
	require.Equal(t, int64(10), idx)
}

func reverseCopy(s []string) []string {
	a := make([]string, len(s))
	for i, j := 0, len(s)-1; i < len(s); i++ {
		a[i] = s[j]
		j--
	}
	return a
}

func TestBatch(t *testing.T) {
	var err error
	// SetContextLogger(NewContextLogger(DebugLevel))

	eds := testFirestore(t)
	ctx := context.TODO()
	path := testPath()
	t.Logf("Path: %s", path)

	values := []events.Document{}
	length := 1001
	for i := 0; i < length; i++ {
		str := fmt.Sprintf("value%d", i)
		values = append(values, dstore.Data([]byte(str)))
	}
	idx, err := eds.EventsAdd(ctx, path, values)
	require.NoError(t, err)
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
		require.Equal(t, fmt.Sprintf("value%d", i), string(event.Data()))
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

	_, err := ds.EventsAdd(ctx, path, []events.Document{
		dstore.Data([]byte("test1")),
		dstore.Data([]byte("test2")),
	})
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
	require.Equal(t, []byte("test1"), event1.Data())
	event2, err := iter.Next()
	require.NoError(t, err)
	require.Equal(t, []byte("test2"), event2.Data())
}

func TestIncrement(t *testing.T) {
	eds := testFirestore(t)
	ctx := context.TODO()
	collection := testCollection()
	path := dstore.Path(collection, "key1")

	n, i, err := eds.Increment(ctx, path, "count", 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	require.Equal(t, int64(1), i)

	n, i, err = eds.Increment(ctx, path, "count", 5)
	require.NoError(t, err)
	require.Equal(t, int64(6), n)
	require.Equal(t, int64(2), i)
}
