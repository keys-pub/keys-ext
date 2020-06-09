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

var fixtures = map[string][]byte{
	"test.txt":  []byte("testdata"),
	"test2.txt": []byte("testdata2"),
}

type closeFn func()

func testConfig(t *testing.T) (syncp.Config, closeFn) {
	tmpDir, err := ioutil.TempDir("", "TestSyncp-"+keys.RandFileName())
	require.NoError(t, err)
	t.Logf("Dir: %s", tmpDir)
	closeFn := func() { os.RemoveAll(tmpDir) }
	cfg := syncp.Config{
		Dir: tmpDir,
	}
	return cfg, closeFn
}

func testProgramSync(t *testing.T, pr syncp.Program, cfg syncp.Config, rt syncp.Runtime) {
	var err error

	path1 := keys.RandFileName() + ".txt"
	path2 := keys.RandFileName() + ".txt"
	testSaveFiles(t, cfg, map[string][]byte{
		path1:     []byte("testdata"),
		path2:     []byte("testdata2"),
		".hidden": []byte("testhidden"),
	})

	err = pr.Sync(cfg, syncp.WithRuntime(rt))
	require.NoError(t, err)

	fileInfos, err := ioutil.ReadDir(cfg.Dir)
	require.NoError(t, err)
	files := fileNames(fileInfos)

	// Check files
	testFile(t, filepath.Join(cfg.Dir, path1), []byte("testdata"), files)
	testFile(t, filepath.Join(cfg.Dir, path2), []byte("testdata2"), files)

	// Test remote fixtures
	for path, b := range fixtures {
		testFile(t, filepath.Join(cfg.Dir, path), b, files)
	}

	// Remove hidden/excluded and re-sync
	hidden := filepath.Join(cfg.Dir, ".hidden")
	err = os.Remove(hidden)
	require.NoError(t, err)
	err = pr.Sync(cfg, syncp.WithRuntime(rt))
	require.NoError(t, err)
	require.NoFileExists(t, hidden)
}

func testFixtures(t *testing.T, pr syncp.Program, cfg syncp.Config) {
	var err error
	testSaveFiles(t, cfg, fixtures)

	rt := newTestRuntime(t)
	err = pr.Sync(cfg, syncp.WithRuntime(rt))
	require.NoError(t, err)

	fileInfos, err := ioutil.ReadDir(cfg.Dir)
	require.NoError(t, err)
	files := fileNames(fileInfos)
	for path, b := range fixtures {
		testFile(t, filepath.Join(cfg.Dir, path), b, files)
	}
}

func testSaveFiles(t *testing.T, cfg syncp.Config, files map[string][]byte) {
	var err error
	for file, data := range files {
		path := filepath.Join(cfg.Dir, file)
		dir, _ := filepath.Split(path)
		err = os.MkdirAll(dir, 0700)
		require.NoError(t, err)
		err = ioutil.WriteFile(path, data, 0600)
		require.NoError(t, err)
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

type testRuntime struct {
	t *testing.T
}

func (l *testRuntime) Log(format string, args ...interface{}) {
	l.t.Logf(format, args...)
}

func newTestRuntime(t *testing.T) syncp.Runtime {
	return &testRuntime{t: t}
}
