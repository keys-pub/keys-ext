package firestore

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/stretchr/testify/require"
)

func TestFirestoreChanges(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	fs := testFirestore(t)
	testChanges(t, fs, fs)
}

func testChanges(t *testing.T, dst ds.DocumentStore, changes ds.Changes) {
	ctx := context.TODO()

	collection := testCollection()
	changesCollection := "changes-" + collection
	paths := []string{}
	length := 40

	for i := 0; i < length; i++ {
		id := fmt.Sprintf("%s-%06d", keys.RandFileName(), i)
		path := ds.Path(collection, id)
		paths = append(paths, path)
		err := dst.Create(ctx, path, []byte(fmt.Sprintf("value%d", i)))
		require.NoError(t, err)
		err = changes.ChangeAdd(ctx, changesCollection, id, path)
		require.NoError(t, err)
	}

	sorted := stringsCopy(paths)
	sort.Strings(sorted)

	iter, err := dst.Documents(ctx, collection, ds.Index(1), ds.Limit(2))
	require.NoError(t, err)
	doc, err := iter.Next()
	require.NoError(t, err)
	require.Equal(t, sorted[1], doc.Path)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Equal(t, sorted[2], doc.Path)
	iter.Release()

	// Changes (limit=10, asc)
	recent, ts, err := changes.Changes(ctx, changesCollection, time.Time{}, 10, ds.Ascending)
	require.NoError(t, err)
	require.Equal(t, 10, len(recent))
	recentPaths := []string{}
	for _, doc := range recent {
		recentPaths = append(recentPaths, doc.Path)
	}
	require.Equal(t, paths[0:10], recentPaths)

	// Changes (ts, asc)
	recent, ts, err = changes.Changes(ctx, changesCollection, ts, 10, ds.Ascending)
	require.NoError(t, err)
	require.False(t, ts.IsZero())
	require.Equal(t, 10, len(recent))
	recentPaths = []string{}
	for _, doc := range recent {
		recentPaths = append(recentPaths, doc.Path)
	}
	require.Equal(t, paths[9:19], recentPaths)

	// Changes (now)
	now := time.Now()
	recent, ts, err = changes.Changes(ctx, changesCollection, now, 100, ds.Ascending)
	require.NoError(t, err)
	require.Equal(t, 0, len(recent))
	require.Equal(t, now, ts)

	// Descending
	revpaths := reverseCopy(paths)

	// Changes (limit=10, desc)
	recent, ts, err = changes.Changes(ctx, changesCollection, time.Time{}, 10, ds.Descending)
	require.NoError(t, err)
	require.Equal(t, 10, len(recent))
	require.False(t, ts.IsZero())
	recentPaths = []string{}
	for _, doc := range recent {
		recentPaths = append(recentPaths, doc.Path)
	}
	require.Equal(t, revpaths[0:10], recentPaths)

	// Changes (limit=5, ts, desc)
	recent, ts, err = changes.Changes(ctx, changesCollection, ts, 5, ds.Descending)
	require.NoError(t, err)
	require.Equal(t, 5, len(recent))
	require.False(t, ts.IsZero())
	recentPaths = []string{}
	for _, doc := range recent {
		recentPaths = append(recentPaths, doc.Path)
	}
	require.Equal(t, revpaths[9:14], recentPaths)
}

func stringsCopy(s []string) []string {
	a := make([]string, len(s))
	copy(a, s)
	return a
}

func reverseCopy(s []string) []string {
	a := make([]string, len(s))
	for i, j := 0, len(s)-1; i < len(s); i++ {
		a[i] = s[j]
		j--
	}
	return a
}
