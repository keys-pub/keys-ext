package service

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestWaitForPID(t *testing.T) {
	path := filepath.Join(os.TempDir(), "test.pid")

	defer os.Remove(path)

	pid, err := waitForPID(path, checkNoop, time.Millisecond, time.Millisecond*100)
	require.EqualError(t, err, "timed out waiting for pid")
	require.Equal(t, -1, pid)

	err = ioutil.WriteFile(path, []byte("abc"), 0600)
	require.NoError(t, err)
	pid2, err := waitForPID(path, checkNoop, time.Millisecond, time.Second)
	require.EqualError(t, err, "strconv.Atoi: parsing \"abc\": invalid syntax")
	require.Equal(t, -1, pid2)

	err = ioutil.WriteFile(path, []byte("-1"), 0600)
	require.NoError(t, err)
	pid3, err := waitForPID(path, checkNoop, time.Millisecond, time.Second)
	require.EqualError(t, err, "negative pid")
	require.Equal(t, -1, pid3)

	err = ioutil.WriteFile(path, []byte("123"), 0600)
	require.NoError(t, err)
	pid4, err := waitForPID(path, checkNoop, time.Millisecond, time.Second)
	require.NoError(t, err)
	require.Equal(t, 123, pid4)

	errFn := func() error { return errors.Errorf("test error") }
	pid5, err := waitForPID(path, errFn, time.Millisecond, time.Second)
	require.EqualError(t, err, "test error")
	require.Equal(t, -1, pid5)

	os.Remove(path)
	errFn = func() error { return errors.Errorf("test error2") }
	pid6, err6 := waitForPID(path, errFn, time.Millisecond, time.Second)
	require.EqualError(t, err6, "test error2")
	require.Equal(t, -1, pid6)
}
