package vault

import (
	"testing"
	"time"

	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestSetGetValue(t *testing.T) {
	var err error

	vlt := New(NewMem())

	err = vlt.setPullIndex(1)
	require.NoError(t, err)
	idx, err := vlt.pullIndex()
	require.NoError(t, err)
	require.Equal(t, int64(1), idx)

	err = vlt.setPushIndex(1)
	require.NoError(t, err)
	idx, err = vlt.pushIndex()
	require.NoError(t, err)
	require.Equal(t, int64(1), idx)
	idx, err = vlt.pushIndexNext()
	require.NoError(t, err)
	require.Equal(t, int64(2), idx)

	n, err := vlt.getInt64("/test/int64")
	require.NoError(t, err)
	require.Equal(t, int64(0), n)
	err = vlt.setInt64("/test/int64", 123)
	require.NoError(t, err)
	n, err = vlt.getInt64("/test/int64")
	require.NoError(t, err)
	require.Equal(t, int64(123), n)

	b, err := vlt.getBool("/test/bool")
	require.NoError(t, err)
	require.Equal(t, false, b)
	err = vlt.setBool("/test/bool", true)
	require.NoError(t, err)
	b, err = vlt.getBool("/test/bool")
	require.NoError(t, err)
	require.Equal(t, true, b)

	tm, err := vlt.getTime("/test/time")
	require.NoError(t, err)
	require.True(t, tm.IsZero())
	now := tsutil.ParseMillis(tsutil.Millis(time.Now()))
	err = vlt.setTime("/test/time", now)
	require.NoError(t, err)
	tm, err = vlt.getTime("/test/time")
	require.NoError(t, err)
	require.Equal(t, now, tm)
}
