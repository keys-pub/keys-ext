package syncp_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/syncp"

	"github.com/stretchr/testify/require"
)

func TestGSUtil(t *testing.T) {
	if os.Getenv("TEST_GSUTIL") != "1" {
		t.Skip()
	}
	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	tmpDir, err := ioutil.TempDir("", "TestGSUtil-"+keys.RandFileName())
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	cfg := syncp.Config{
		Dir: tmpDir,
	}

	existing := map[string][]byte{
		"test.txt":  []byte("testdata"),
		"test2.txt": []byte("testdata2"),
	}

	gsutil, err := syncp.NewGSUtil("gs://keys-chill-test")
	require.NoError(t, err)

	res := testProgramSync(t, gsutil, cfg, existing)
	require.NoError(t, res.Err)
	require.Equal(t, 2, len(res.CmdResults))
	cmd0 := res.CmdResults[0].Cmd
	cmd1 := res.CmdResults[1].Cmd

	require.NotEmpty(t, cmd0.BinPath)
	expectedArgs := []string{"-m", "rsync", "-e", "-x", "\\.git$", cfg.Dir, "gs://keys-chill-test"}
	require.Equal(t, expectedArgs, cmd0.Args)
	require.NotEmpty(t, cmd1.BinPath)
	expectedArgs2 := []string{"-m", "rsync", "-e", "-x", "\\.git$", "gs://keys-chill-test", cfg.Dir}
	require.Equal(t, expectedArgs2, cmd1.Args)
}

func TestGSUtilValidate(t *testing.T) {
	_, err := syncp.NewGSUtil("keys-chill-test")
	require.EqualError(t, err, "invalid bucket scheme, expected gs://")
}
