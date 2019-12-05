package server

import (
	"context"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestMemTestCache(t *testing.T) {
	clock := newClock()
	mc := newMemTestCache(clock.Now)

	n1 := keys.RandID().String()
	val, err := mc.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Empty(t, val)

	err = mc.Set(context.TODO(), n1, "1")
	require.NoError(t, err)

	val2, err := mc.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.NotEmpty(t, val2)

	n, err := mc.Increment(context.TODO(), n1)
	require.NoError(t, err)
	require.Equal(t, 2, n)
}

func TestMemTestCacheExpiration(t *testing.T) {
	// SetLog(newLog(DebugLevel))
	clock := newClock()
	clock.setTick(time.Second)
	mc := NewMemTestCache(clock.Now)

	n1 := keys.RandID().String()
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

	n2 := keys.RandID().String()
	err = mc.Set(context.TODO(), n2, "2")
	require.NoError(t, err)
	err = mc.Expire(context.TODO(), n2, time.Minute)
	require.NoError(t, err)

	val3, err := mc.Get(context.TODO(), n2)
	require.NoError(t, err)
	require.NotEmpty(t, val3)
}
