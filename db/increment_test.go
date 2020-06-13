package db_test

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
	d, closeFn := testDBWithOpts(t, path, key)
	defer closeFn()

	prev := ""
	for i := 0; i < 100000; i++ {
		n, err := d.Increment(ctx, ds.Path("db/increment"))
		require.NoError(t, err)
		require.True(t, n > prev)
		prev = n
	}
	t.Logf("Prev: %s", prev)
	d.Close()

	// Re-open and do increment
	d2, closeFn2 := testDBWithOpts(t, path, key)
	defer closeFn2()

	for i := 0; i < 100000; i++ {
		n, err := d2.Increment(ctx, ds.Path("db/increment"))
		require.NoError(t, err)
		require.True(t, n > prev)
		prev = n
	}
	t.Logf("Prev (2): %s", prev)

}
