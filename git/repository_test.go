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

func TestRepositoryAddDelete(t *testing.T) {
	git.SetLogger(git.NewLogger(git.DebugLevel))

	path := keys.RandTempPath("")
	path2 := keys.RandTempPath("")

	privateKey, err := ioutil.ReadFile("id_ed25519")
	require.NoError(t, err)
	key, err := keys.ParseSSHKey(privateKey, nil, true)
	require.NoError(t, err)

	url := "git@gitlab.com:gabrielha/pass.git"
	host := "gitlab.com"

	ks := keys.NewMemStore(true)
	err = ks.Save(key)
	require.NoError(t, err)

	// Repo1: Open
	repo, err := git.NewRepository(url, host, path, key, nil)
	require.NoError(t, err)
	err = repo.Open()
	require.NoError(t, err)

	// Repo2: Open
	repo2, err := git.NewRepository(url, host, path2, key, nil)
	require.NoError(t, err)
	err = repo2.Open()
	require.NoError(t, err)

	item := keyring.NewItem(keys.Rand3262(), []byte("mypassword"), "", time.Now())

	// Repo1: Add, Push
	err = repo.Add(item)
	require.NoError(t, err)
	err = repo.Push()
	require.NoError(t, err)

	// Repo3: Open (Repo1 path)
	repo3, err := git.NewRepository(url, host, path, key, nil)
	require.NoError(t, err)
	err = repo3.Open()
	require.NoError(t, err)
	items, err := repo2.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(items))
	require.Equal(t, []byte("mypassword"), items[0].Data)

	// Repo2: Pull, List
	err = repo2.Pull()
	require.NoError(t, err)
	items, err = repo2.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(items))
	require.Equal(t, []byte("mypassword"), items[0].Data)

	// Repo2: Delete, Push
	err = repo2.Delete(item.ID)
	require.NoError(t, err)
	err = repo2.Push()
	require.NoError(t, err)

	// Repo1: Pull
	err = repo.Pull()
	require.NoError(t, err)
	items, err = repo.List()
	require.NoError(t, err)
	require.Equal(t, 0, len(items))
}

func TestRepositoryConflict(t *testing.T) {
	git.SetLogger(git.NewLogger(git.DebugLevel))

	path := keys.RandTempPath("")
	t.Logf("Path: %s", path)
	path2 := keys.RandTempPath("")
	t.Logf("Path2: %s", path2)

	privateKey, err := ioutil.ReadFile("id_ed25519")
	require.NoError(t, err)
	key, err := keys.ParseSSHKey(privateKey, nil, true)
	require.NoError(t, err)

	url := "git@gitlab.com:gabrielha/pass.git"
	host := "gitlab.com"

	// Repo1: Open
	repo, err := git.NewRepository(url, host, path, key, nil)
	require.NoError(t, err)
	err = repo.Open()
	require.NoError(t, err)

	// Repo2: Open
	repo2, err := git.NewRepository(url, host, path2, key, nil)
	require.NoError(t, err)
	err = repo2.Open()
	require.NoError(t, err)

	item := keyring.NewItem(keys.Rand3262(), []byte("mypassword"), "", time.Now())

	// Repo1: Add, Push
	err = repo.Add(item)
	require.NoError(t, err)
	err = repo.Push()
	require.NoError(t, err)

	// Repo2: Pull, List
	err = repo2.Pull()
	require.NoError(t, err)
	items, err := repo2.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(items))
	require.Equal(t, []byte("mypassword"), items[0].Data)

	// Repo1: Add, Push
	item.Data = []byte("mypassword2")
	err = repo.Add(item)
	require.NoError(t, err)
	err = repo.Push()
	require.NoError(t, err)

	// Repo2: Add, Push
	item.Data = []byte("mypassword3")
	err = repo2.Add(item)
	require.NoError(t, err)
	err = repo2.Push()
	require.EqualError(t, err, "failed to push: cannot push because a reference that you are trying to update on the remote contains commits that are not present locally.")

	// Repo1: Delete, Push
	err = repo.Delete(item.ID)
	require.NoError(t, err)
	err = repo.Push()
	require.NoError(t, err)
}
