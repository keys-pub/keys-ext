package syncp_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/syncp"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
	if os.Getenv("TEST_GIT") != "1" {
		t.Skip()
	}
	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	tmpDir, err := ioutil.TempDir("", "TestGit-"+keys.RandFileName())
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	cfg := syncp.Config{
		Dir: tmpDir,
	}
	t.Logf("Dir: %s", tmpDir)

	existing := map[string][]byte{
		"test.txt":  []byte("testdata"),
		"test2.txt": []byte("testdata2"),
	}

	repo := "git@gitlab.com:gabrielha/keys.pub.test.git"
	git, err := syncp.NewGit(repo)
	require.NoError(t, err)

	// Setup
	func() {
		res := git.Setup(cfg)
		require.NoError(t, res.Err)
		t.Logf("%s", res)

		require.Equal(t, 2, len(res.CmdResults))
		cmd0 := res.CmdResults[0].Cmd
		cmd1 := res.CmdResults[1].Cmd

		require.NotEmpty(t, cmd0.BinPath)
		expectedArgs := []string{"init"}
		require.Equal(t, expectedArgs, cmd0.Args)

		require.NotEmpty(t, cmd1.BinPath)
		expectedArgs2 := []string{"remote", "add", "origin", git.Remote()}
		require.Equal(t, expectedArgs2, cmd1.Args)
	}()

	// Sync
	func() {
		res := testProgramSync(t, git, cfg, existing)
		t.Logf("%s", res)

		require.Equal(t, 4, len(res.CmdResults))
		cmd0 := res.CmdResults[0].Cmd
		cmd1 := res.CmdResults[1].Cmd
		cmd2 := res.CmdResults[2].Cmd
		cmd3 := res.CmdResults[3].Cmd

		require.NotEmpty(t, cmd0.BinPath)
		expectedArgs := []string{"pull", "origin", "master"}
		require.Equal(t, expectedArgs, cmd0.Args)

		require.NotEmpty(t, cmd1.BinPath)
		expectedArgs = []string{"add", "."}
		require.Equal(t, expectedArgs, cmd1.Args)

		require.NotEmpty(t, cmd2.BinPath)
		expectedArgs = []string{"commit", "-m", "Syncing..."}
		require.Equal(t, expectedArgs, cmd2.Args)

		require.NotEmpty(t, cmd3.BinPath)
		expectedArgs = []string{"push", "origin", "master"}
		require.Equal(t, expectedArgs, cmd3.Args)
	}()
}
