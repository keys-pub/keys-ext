package syncp_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/syncp"
	"github.com/stretchr/testify/require"
)

type fixture map[string][]byte

func testProgramSync(t *testing.T, pr syncp.Program, cfg syncp.Config, rt syncp.Runtime, existing fixture) {
	var err error

	// Write test files
	path1 := keys.RandFileName() + ".txt"
	err = ioutil.WriteFile(filepath.Join(cfg.Dir, path1), []byte("testdata"), 0600)
	require.NoError(t, err)
	path2 := keys.RandFileName() + ".txt"
	err = ioutil.WriteFile(filepath.Join(cfg.Dir, path2), []byte("testdata2"), 0600)
	require.NoError(t, err)

	err = pr.Sync(cfg, rt)
	require.NoError(t, err)

	fileInfos, err := ioutil.ReadDir(cfg.Dir)
	require.NoError(t, err)
	files := fileNames(fileInfos)

	// Check test files
	testFile(t, filepath.Join(cfg.Dir, path1), []byte("testdata"), files)
	testFile(t, filepath.Join(cfg.Dir, path2), []byte("testdata2"), files)

	// Test existing files
	for path, b := range existing {
		testFile(t, filepath.Join(cfg.Dir, path), b, files)
	}
}

func testFile(t *testing.T, path string, expected []byte, files []string) {
	require.Contains(t, files, filepath.Base(path))
	require.FileExists(t, path)
	b, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, expected, b)
}

func fileNames(fs []os.FileInfo) []string {
	names := make([]string, 0, len(fs))
	for _, f := range fs {
		names = append(names, f.Name())
	}
	return names
}
