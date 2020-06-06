package sync_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/sync"
	"github.com/stretchr/testify/require"
)

func testGit(t *testing.T, cfg sync.Config) sync.Program {
	git, err := sync.NewGit()
	require.NoError(t, err)
	cmds, err := git.Commands(cfg)
	require.NoError(t, err)

	require.Equal(t, 4, len(cmds))
	require.NotEmpty(t, cmds[0].BinPath)
	expectedArgs := []string{"pull", "origin", "master"}
	require.Equal(t, expectedArgs, cmds[0].Args)

	require.NotEmpty(t, cmds[1].BinPath)
	expectedArgs = []string{"add", "."}
	require.Equal(t, expectedArgs, cmds[1].Args)

	require.NotEmpty(t, cmds[2].BinPath)
	expectedArgs = []string{"commit", "-m", "Syncing..."}
	require.Equal(t, expectedArgs, cmds[2].Args)

	require.NotEmpty(t, cmds[3].BinPath)
	expectedArgs = []string{"push", "origin", "master"}
	require.Equal(t, expectedArgs, cmds[3].Args)
	return git
}

func testGitSetup(t *testing.T, cfg sync.Config) sync.Program {
	repo := "git@gitlab.com:gabrielha/keys.pub.test.git"
	git, err := sync.NewGitSetup(repo)
	require.NoError(t, err)
	cmds, err := git.Commands(cfg)
	require.NoError(t, err)
	require.Equal(t, 2, len(cmds))

	require.NotEmpty(t, cmds[0].BinPath)
	expectedArgs := []string{"init"}
	require.Equal(t, expectedArgs, cmds[0].Args)

	require.NotEmpty(t, cmds[1].BinPath)
	expectedArgs2 := []string{"remote", "add", "origin", repo}
	require.Equal(t, expectedArgs2, cmds[1].Args)
	return git
}

func TestGit(t *testing.T) {
	if os.Getenv("TEST_GIT") != "1" {
		t.Skip()
	}
	sync.SetLogger(sync.NewLogger(sync.DebugLevel))

	tmpDir, err := ioutil.TempDir("", "TestGit-"+keys.RandFileName())
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	cfg := sync.Config{
		Dir: tmpDir,
	}
	t.Logf("Dir: %s", tmpDir)

	existing := map[string][]byte{
		"test.txt":  []byte("testdata"),
		"test2.txt": []byte("testdata2"),
	}

	// Setup
	setup := testGitSetup(t, cfg)
	err = sync.Run(setup, cfg)
	require.NoError(t, err)

	pr := testGit(t, cfg)
	testProgram(t, pr, cfg, existing)
}
