package sdb_test

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/stretchr/testify/require"
)

func TestIncrement(t *testing.T) {
	// db.SetLogger(db.NewLogger(db.DebugLevel))

	ctx := context.TODO()
	path := testPath()
	key := keys.Rand32()
	db, closeFn := testDBWithOpts(t, path, key)
	defer closeFn()

	prev := ""
	for i := 0; i < 100000; i++ {
		n, err := db.Increment(ctx, ds.Path("db/increment"))
		require.NoError(t, err)
		require.True(t, n > prev)
		prev = n
	}
	// t.Logf("Prev: %s", prev)
	db.Close()

	// Re-open and do increment
	db2, closeFn2 := testDBWithOpts(t, path, key)
	defer closeFn2()

	for i := 0; i < 100000; i++ {
		n, err := db2.Increment(ctx, ds.Path("db/increment"))
		require.NoError(t, err)
		require.True(t, n > prev)
		prev = n
	}
	// t.Logf("Prev (2): %s", prev)
}
