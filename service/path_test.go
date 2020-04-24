package service

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIncPath(t *testing.T) {
	require.Equal(t, "test.txt", incPath("test.txt", 0))
	require.Equal(t, "test-1.txt", incPath("test.txt", 1))
	require.Equal(t, "test-2.txt", incPath("test.txt", 2))
}

func TestNextPath(t *testing.T) {
	p := filepath.Join(os.TempDir(), "next.txt")

	out, err := nextPathIfExists(p)
	require.NoError(t, err)
	require.Equal(t, p, out)

	err = ioutil.WriteFile(p, []byte("1"), 0644)
	require.NoError(t, err)
	defer os.Remove(p)

	out, err = nextPathIfExists(p)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(os.TempDir(), "next-1.txt"), out)

	err = ioutil.WriteFile(out, []byte("2"), 0644)
	require.NoError(t, err)
	defer os.Remove(out)

	out, err = nextPathIfExists(p)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(os.TempDir(), "next-2.txt"), out)
}
