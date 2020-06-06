package sync_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/keys-pub/keys"

	"github.com/keys-pub/keys-ext/sync"
	"github.com/stretchr/testify/require"
)

func TestGSUtil(t *testing.T) {
	if os.Getenv("TEST_GSUTIL") != "1" {
		t.Skip()
	}
	sync.SetLogger(sync.NewLogger(sync.DebugLevel))

	tmpDir, err := ioutil.TempDir("", "TestGSUtil-"+keys.RandFileName())
	require.NoError(t, err)
	// defer os.RemoveAll(tmpDir)
	cfg := sync.Config{
		Dir: tmpDir,
	}

	existing := map[string][]byte{
		"test.txt":  []byte("testdata"),
		"test2.txt": []byte("testdata2"),
	}

	pr := testGSUtil(t, cfg)
	testProgram(t, pr, cfg, existing)
}

func testGSUtil(t *testing.T, cfg sync.Config) sync.Program {
	gsutil, err := sync.NewGSUtil("", "keys-chill-test")
	require.NoError(t, err)

	cmds, err := gsutil.Commands(cfg)
	require.NoError(t, err)
	require.Equal(t, 2, len(cmds))
	require.NotEmpty(t, cmds[0].BinPath)
	expectedArgs := []string{"-m", "rsync", "-e", "-x", "\\.git$", cfg.Dir, "gs://keys-chill-test"}
	require.Equal(t, expectedArgs, cmds[0].Args)
	require.NotEmpty(t, cmds[1].BinPath)
	expectedArgs2 := []string{"-m", "rsync", "-e", "-x", "\\.git$", "gs://keys-chill-test", cfg.Dir}
	require.Equal(t, expectedArgs2, cmds[1].Args)
	return gsutil
}
