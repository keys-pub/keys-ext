package vault

import (
	"testing"
	"time"

	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	var err error

	vlt := New(NewMem())

	err = vlt.setIndex(1)
	idx, err := vlt.index()
	require.NoError(t, err)
	require.Equal(t, int64(1), idx)

	n, err := vlt.getConfigInt64("testint64")
	require.NoError(t, err)
	require.Equal(t, int64(0), n)
	err = vlt.setConfigInt64("testint64", 123)
	require.NoError(t, err)
	n, err = vlt.getConfigInt64("testint64")
	require.Equal(t, int64(123), n)

	b, err := vlt.getConfigBool("testbool")
	require.NoError(t, err)
	require.Equal(t, false, b)
	err = vlt.setConfigBool("testbool", true)
	require.NoError(t, err)
	b, err = vlt.getConfigBool("testbool")
	require.Equal(t, true, b)

	tm, err := vlt.getConfigTime("testtime")
	require.NoError(t, err)
	require.True(t, tm.IsZero())
	now := tsutil.ParseMillis(tsutil.Millis(time.Now()))
	err = vlt.setConfigTime("testtime", now)
	require.NoError(t, err)
	tm, err = vlt.getConfigTime("testtime")
	require.Equal(t, now, tm)
}
