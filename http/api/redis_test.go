package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestRedis(t *testing.T) {
	clock := tsutil.NewTestClock()
	rds := api.NewRedisTest(clock)

	n1 := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	val, err := rds.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Empty(t, val)

	err = rds.Set(context.TODO(), n1, "1")
	require.NoError(t, err)

	val, err = rds.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Equal(t, "1", val)

	n, err := rds.Increment(context.TODO(), n1)
	require.NoError(t, err)
	require.Equal(t, int64(2), n)

	val, err = rds.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Equal(t, "2", val)
}

func TestRedisExpiration(t *testing.T) {
	// SetLog(newLog(DebugLevel))
	clock := tsutil.NewTestClock()
	rds := api.NewRedisTest(clock)

	n1 := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	val, err := rds.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Empty(t, val)

	err = rds.Set(context.TODO(), n1, "1")
	require.NoError(t, err)
	err = rds.Expire(context.TODO(), n1, time.Millisecond)
	require.NoError(t, err)

	val2, err := rds.Get(context.TODO(), n1)
	require.NoError(t, err)
	require.Empty(t, val2)

	n2 := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	err = rds.Set(context.TODO(), n2, "2")
	require.NoError(t, err)
	err = rds.Expire(context.TODO(), n2, time.Minute)
	require.NoError(t, err)

	val3, err := rds.Get(context.TODO(), n2)
	require.NoError(t, err)
	require.Equal(t, "2", val3)
}
