package git_test

import (
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

	// Repo
	repo, err := git.NewRepository(url, host, path, key, nil)
	require.NoError(t, err)
	err = repo.Open()
	require.NoError(t, err)

	// Keyring
	kr := keyring.NewMem(true)
	item := keyring.NewItem(keys.Rand3262(), []byte("testpassword"), "", time.Now())
	err = kr.Create(item)
	require.NoError(t, err)

	changes, err := git.Export(kr, repo)
	require.NoError(t, err)
	require.Equal(t, 1, len(changes.Add))
	require.Equal(t, item.ID, changes.Add[0].ID)
}
