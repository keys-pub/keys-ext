package firestore

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/stretchr/testify/require"
)

func TestFirestoreChanges(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	fs := testFirestore(t, true)
	testChanges(t, fs, fs)
}

func genPath(n int) string {
	return encoding.MustEncode(keys.SHA256([]byte{byte(n)}), encoding.Base62)
}

func testChanges(t *testing.T, ds keys.DocumentStore, changes keys.Changes) {
	ctx := context.TODO()

	length := 40
	paths := []string{}
	for i := 0; i < length; i++ {
		p := keys.Path("test", fmt.Sprintf("%s-%06d", genPath(i), i))
		paths = append(paths, p)
	}

	for i, p := range paths {
		err := ds.Create(ctx, p, []byte(fmt.Sprintf("value%d", i)))
		require.NoError(t, err)
		err = changes.ChangeAdd(ctx, "test-changes", p)
		require.NoError(t, err)
		change, err := changes.Change(ctx, "test-changes", p)
		require.NoError(t, err)
		require.Equal(t, p, change.Path)
		require.True(t, !change.Timestamp.IsZero())
	}

	sorted := stringsCopy(paths)
	sort.Strings(sorted)

	iter, err := ds.Documents(ctx, "test", &keys.DocumentsOpts{Index: 1, Limit: 2})
	require.NoError(t, err)
	doc, err := iter.Next()
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, sorted[1], doc.Path)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, sorted[2], doc.Path)
	iter.Release()

	// Changes (limit=10, asc)
	recent, ts, err := changes.Changes(ctx, "test-changes", time.Time{}, 10, keys.Ascending)
	require.NoError(t, err)
	require.Equal(t, 10, len(recent))
	recentPaths := []string{}
	for _, doc := range recent {
		recentPaths = append(recentPaths, doc.Path)
	}
	require.Equal(t, paths[0:10], recentPaths)

	// Changes (ts, asc)
	recent, ts, err = changes.Changes(ctx, "test-changes", ts, 10, keys.Ascending)
	require.NoError(t, err)
	require.False(t, ts.IsZero())
	require.Equal(t, 10, len(recent))
	recentPaths = []string{}
	for _, doc := range recent {
		recentPaths = append(recentPaths, doc.Path)
	}
	require.Equal(t, paths[9:19], recentPaths)

	revpaths := reverseCopy(paths)

	// Changes (limit=10, desc)
	recent, ts, err = changes.Changes(ctx, "test-changes", time.Time{}, 10, keys.Descending)
	require.NoError(t, err)
	require.Equal(t, 10, len(recent))
	require.False(t, ts.IsZero())
	recentPaths = []string{}
	for _, doc := range recent {
		recentPaths = append(recentPaths, doc.Path)
	}
	require.Equal(t, revpaths[0:10], recentPaths)

	// Changes (limit=5, ts, desc)
	recent, ts, err = changes.Changes(ctx, "test-changes", ts, 5, keys.Descending)
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
