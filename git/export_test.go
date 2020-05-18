package git_test

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keysd/git"
	"github.com/stretchr/testify/require"
)

func TestExport(t *testing.T) {
	git.SetLogger(git.NewLogger(git.DebugLevel))

	path := keys.RandTempPath("")

	privateKey, err := ioutil.ReadFile("id_ed25519")
	require.NoError(t, err)
	key, err := keys.ParseSSHKey(privateKey, nil, true)
	require.NoError(t, err)

	url := "git@gitlab.com:gabrielha/pass.git"
	host := "gitlab.com"

	service := "GitTest-Export-" + keys.Rand3262()

	salt := bytes.Repeat([]byte{0x01}, 16)
	auth, err := keyring.NewPasswordAuth("testpassword", salt)
	require.NoError(t, err)

	// Repo
	repo, err := git.NewRepository(url, host, path, key, nil)
	require.NoError(t, err)
	err = repo.Open()
	require.NoError(t, err)

	// Keyring #1 (mem)
	kr := keyring.NewMem(true)
	item := keyring.NewItem(keys.Rand3262(), []byte("testpassword"), "", time.Now())
	err = kr.Create(item)
	require.NoError(t, err)

	// Keyring #2 (git repo)
	kr2, err := keyring.New(service, repo)
	require.NoError(t, err)
	err = kr2.Unlock(auth)
	require.NoError(t, err)
	err = repo.Push()
	require.NoError(t, err)

	// Export
	changes, err := keyring.Export(kr, kr2)
	require.NoError(t, err)
	require.Equal(t, 1, len(changes.Add))
	require.Equal(t, item.ID, changes.Add[0].ID)
}
