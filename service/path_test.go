package service

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestIncPath(t *testing.T) {
	require.Equal(t, "test.txt", incPath("test.txt", 0))
	require.Equal(t, "test-2.txt", incPath("test.txt", 1))
	require.Equal(t, "test-3.txt", incPath("test.txt", 2))
}

func TestNextPath(t *testing.T) {
	file := keys.RandFileName()
	p := filepath.Join(os.TempDir(), file+".txt")
	defer func() { _ = os.Remove(p) }()

	out, err := nextPathIfExists(p)
	require.NoError(t, err)
	defer func() { _ = os.Remove(out) }()
	require.Equal(t, p, out)

	err = ioutil.WriteFile(p, []byte("1"), 0644)
	require.NoError(t, err)

	out, err = nextPathIfExists(p)
	require.NoError(t, err)
	defer func() { _ = os.Remove(out) }()
	require.Equal(t, filepath.Join(os.TempDir(), file+"-2.txt"), out)

	err = ioutil.WriteFile(out, []byte("2"), 0644)
	require.NoError(t, err)

	out, err = nextPathIfExists(p)
	require.NoError(t, err)
	defer func() { _ = os.Remove(out) }()
	require.Equal(t, filepath.Join(os.TempDir(), file+"-3.txt"), out)
}

func TestResolveOutPath(t *testing.T) {
	var out string
	var err error

	dir := os.TempDir()
	testFile := filepath.Join(dir, "test.file")
	err = ioutil.WriteFile(testFile, []byte{0x01}, filePerms)
	require.NoError(t, err)
	defer func() { _ = os.Remove(testFile) }()

	out, err = resolveOutPath("", "file.enc", ".enc")
	require.NoError(t, err)
	require.Equal(t, "file", out)

	out, err = resolveOutPath("", testFile, "")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "test-2.file"), out)

	out, err = resolveOutPath("", "file.signed", ".signed")
	require.NoError(t, err)
	require.Equal(t, "file", out)

	out, err = resolveOutPath(dir, "file.enc", ".enc")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "file"), out)

	out, err = resolveOutPath(filepath.Join(dir, "file2"), "file.enc", ".enc")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "file2"), out)

	out, err = resolveOutPath("file2", filepath.Join(dir, "file"), "")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "file2"), out)
}
