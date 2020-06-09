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

	repo := "git@gitlab.com:gabrielha/keys.pub.test.git"
	git, err := syncp.NewGit(repo)
	require.NoError(t, err)

	rt := syncp.NewRuntime()
	// Setup
	func() {
		err := git.Setup(cfg, rt)
		require.NoError(t, err)
	}()

	// Sync
	func() {
		existing := map[string][]byte{
			"test.txt":  []byte("testdata"),
			"test2.txt": []byte("testdata2"),
		}

		testProgramSync(t, git, cfg, rt, existing)
	}()

	// t.Logf(strings.Join(rt.Logs(), "\n"))
}
