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

func TestRepositoryAddDelete(t *testing.T) {
	// git.SetLogger(git.NewLogger(git.DebugLevel))

	path := keys.RandTempPath("")
	path2 := keys.RandTempPath("")

	privateKey, err := ioutil.ReadFile("id_ed25519")
	require.NoError(t, err)
	key, err := keys.ParseSSHKey(privateKey, nil, true)
	require.NoError(t, err)

	url := "git@gitlab.com:gabrielha/pass-test.git"

	service := "GitTest-" + keys.Rand3262()

	salt := bytes.Repeat([]byte{0x01}, 16)
	auth, err := keyring.NewPasswordAuth("testpassword", salt)
	require.NoError(t, err)

	// Keyring #1
	repo1, err := git.NewRepository(url, path, key, nil)
	require.NoError(t, err)
	defer repo1.Close()
	err = repo1.Open()
	require.NoError(t, err)
	kr1, err := keyring.New(service, repo1)
	require.NoError(t, err)
	_, err = kr1.Setup(auth)
	require.NoError(t, err)
	err = repo1.Push()
	require.NoError(t, err)

	// Keyring #2
	repo2, err := git.NewRepository(url, path2, key, nil)
	require.NoError(t, err)
	defer repo2.Close()
	err = repo2.Open()
	require.NoError(t, err)
	kr2, err := keyring.New(service, repo2)
	require.NoError(t, err)
	_, err = kr2.Unlock(keyring.NewKeyAuth(keys.Rand32()))
	require.EqualError(t, err, "invalid keyring auth")
	_, err = kr2.Unlock(auth)
	require.NoError(t, err)

	// Repo1: Create, Push
	item := keyring.NewItem(keys.Rand3262(), []byte("mypassword"), "", time.Now())
	err = kr1.Create(item)
	require.NoError(t, err)
	err = repo1.Push()
	require.NoError(t, err)

	// Keyring #3 (same dir as #1)
	repo3, err := git.NewRepository(url, path, key, nil)
	require.NoError(t, err)
	defer repo3.Close()
	err = repo3.Open()
	require.NoError(t, err)
	kr3, err := keyring.New(service, repo3)
	require.NoError(t, err)
	_, err = kr3.Unlock(auth)
	require.NoError(t, err)
	items, err := kr3.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(items))
	require.Equal(t, []byte("mypassword"), items[0].Data)

	// Repo2: Pull, List
	err = repo2.Pull()
	require.NoError(t, err)
	items, err = kr2.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(items))
	require.Equal(t, []byte("mypassword"), items[0].Data)

	// Repo2: Delete, Push
	ok, err := kr2.Delete(item.ID)
	require.NoError(t, err)
	require.True(t, ok)
	err = repo2.Push()
	require.NoError(t, err)

	// Repo1: Pull, List
	err = repo1.Pull()
	require.NoError(t, err)
	items, err = kr1.List()
	require.NoError(t, err)
	require.Equal(t, 0, len(items))
}

func TestConflictResolve(t *testing.T) {
	// git.SetLogger(git.NewLogger(git.DebugLevel))

	path := keys.RandTempPath("")
	path2 := keys.RandTempPath("")

	privateKey, err := ioutil.ReadFile("id_ed25519")
	require.NoError(t, err)
	key, err := keys.ParseSSHKey(privateKey, nil, true)
	require.NoError(t, err)

	url := "git@gitlab.com:gabrielha/pass-test.git"

	service := "GitTest-" + keys.Rand3262()

	salt := bytes.Repeat([]byte{0x01}, 16)
	auth, err := keyring.NewPasswordAuth("testpassword", salt)
	require.NoError(t, err)

	// Keyring #1
	repo1, err := git.NewRepository(url, path, key, nil)
	require.NoError(t, err)
	defer repo1.Close()
	err = repo1.Open()
	require.NoError(t, err)
	kr1, err := keyring.New(service, repo1)
	require.NoError(t, err)
	_, err = kr1.Setup(auth)
	require.NoError(t, err)
	err = repo1.Push()
	require.NoError(t, err)

	// Keyring #2
	repo2, err := git.NewRepository(url, path2, key, nil)
	require.NoError(t, err)
	defer repo2.Close()
	err = repo2.Open()
	require.NoError(t, err)
	kr2, err := keyring.New(service, repo2)
	require.NoError(t, err)
	_, err = kr2.Unlock(auth)
	require.NoError(t, err)

	// Repo1: Create, Push
	item := keyring.NewItem(keys.Rand3262(), []byte("testpassword"), "", time.Now())
	err = kr1.Create(item)
	require.NoError(t, err)
	err = repo1.Push()
	require.NoError(t, err)

	// Repo2: Create, Push (conflict)
	item.Data = []byte("testpassword2")
	err = kr2.Create(item)
	require.NoError(t, err)
	err = repo2.Push()
	require.True(t, git.ErrIsCode(err, git.ErrNonFastForward))

	err = repo2.Pull()
	require.NoError(t, err)
	err = repo2.Push()
	require.NoError(t, err)

	// Repo1: Pull
	err = repo1.Pull()
	require.NoError(t, err)
	out, err := kr2.Get(item.ID)
	require.NoError(t, err)
	require.Equal(t, "testpassword2", string(out.Data))
}
