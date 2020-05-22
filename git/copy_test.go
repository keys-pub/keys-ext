package git_test

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keysd/git"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	// git.SetLogger(git.NewLogger(git.DebugLevel))
	var err error

	service := "GitTest-Export-" + keys.Rand3262()
	url := "git@gitlab.com:gabrielha/pass-test.git"
	privateKey, err := ioutil.ReadFile("id_ed25519")
	require.NoError(t, err)
	repoKey, err := keys.ParseSSHKey(privateKey, nil, true)
	require.NoError(t, err)

	// Keyring #1 (mem)
	kr := keyring.NewMem(false)
	_, err = kr.UnlockWithPassword("testkeyringpassword", true)
	require.NoError(t, err)

	item := keyring.NewItem(keys.Rand3262(), []byte("testpassword"), "", time.Now())
	err = kr.Create(item)
	require.NoError(t, err)

	// Repo #2, Keyring #2
	path2 := keys.RandTempPath("")
	repo2, err := git.NewRepository(url, path2, repoKey, nil)
	require.NoError(t, err)
	err = repo2.Open()
	require.NoError(t, err)
	kr2, err := keyring.New(service, repo2)
	require.NoError(t, err)

	// Copy #1 to #2
	ids, err := keyring.Copy(kr, kr2)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(ids[0], "#auth-"))
	require.Equal(t, []string{"#salt", item.ID}, ids[1:])

	err = repo2.Push()
	require.NoError(t, err)

	// Repo #3, Keyring #3
	path3 := keys.RandTempPath("")
	repo3, err := git.NewRepository(url, path3, repoKey, nil)
	require.NoError(t, err)
	err = repo3.Open()
	require.NoError(t, err)
	kr3, err := keyring.New(service, repo3)
	require.NoError(t, err)
	_, err = kr3.UnlockWithPassword("testkeyringpassword", false)
	require.NoError(t, err)
	out, err := kr3.Get(item.ID)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "testpassword", string(out.Data))
}
