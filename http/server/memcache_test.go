package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/stretchr/testify/require"
)

func TestMemTestCache(t *testing.T) {
	clock := tsutil.NewClock()
	mc := server.NewMemTestCache(clock.Now)

	n1 := keys.Rand3262()
	val, err := mc.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Empty(t, val)

	err = mc.Set(context.TODO(), n1, "1")
	require.NoError(t, err)

	val, err = mc.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Equal(t, "1", val)

	n, err := mc.Increment(context.TODO(), n1)
	require.NoError(t, err)
	require.Equal(t, int64(2), n)

	val, err = mc.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Equal(t, "2", val)
}

func TestMemTestCacheExpiration(t *testing.T) {
	// SetLog(newLog(DebugLevel))
	clock := tsutil.NewClock()
	clock.SetTick(time.Second)
	mc := server.NewMemTestCache(clock.Now)

	n1 := keys.Rand3262()
	val, err := mc.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Empty(t, val)

	err = mc.Set(context.TODO(), n1, "1")
	require.NoError(t, err)
	err = mc.Expire(context.TODO(), n1, time.Millisecond)
	require.NoError(t, err)

	val2, err := mc.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Empty(t, val2)

	n2 := keys.Rand3262()
	err = mc.Set(context.TODO(), n2, "2")
	require.NoError(t, err)
	err = mc.Expire(context.TODO(), n2, time.Minute)
	require.NoError(t, err)

	val3, err := mc.Get(context.TODO(), n2)
	require.NoError(t, err)
	require.NotEmpty(t, val3)
}
