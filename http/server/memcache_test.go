package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/server"
	"github.com/stretchr/testify/require"
)

func TestMemTestCache(t *testing.T) {
	clock := newClock()
	mc := server.NewMemTestCache(clock.Now)

	n1 := keys.RandIDString()
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
	clock := newClock()
	clock.setTick(time.Second)
	mc := server.NewMemTestCache(clock.Now)

	n1 := keys.RandIDString()
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

	n2 := keys.RandIDString()
	err = mc.Set(context.TODO(), n2, "2")
	require.NoError(t, err)
	err = mc.Expire(context.TODO(), n2, time.Minute)
	require.NoError(t, err)

	val3, err := mc.Get(context.TODO(), n2)
	require.NoError(t, err)
	require.NotEmpty(t, val3)
}

func TestMemTestCachePubSub(t *testing.T) {
	clock := newClock()
	mc := server.NewMemTestCache(clock.Now)

	err := mc.Publish(context.TODO(), "key1", "ping")
	require.NoError(t, err)

	ch, err := mc.Subscribe(context.TODO(), "key1")
	require.NoError(t, err)

	err = mc.Publish(context.TODO(), "key1", "ping")
	require.NoError(t, err)

	b1 := <-ch
	require.Equal(t, "ping", string(b1))
	b2 := <-ch
	require.Equal(t, "ping", string(b2))
}
