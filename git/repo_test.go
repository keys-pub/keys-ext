package git_test

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/git"
	"github.com/keys-pub/keys/keyring"
	"github.com/stretchr/testify/require"
)

func TestRepositoryAddDelete(t *testing.T) {
	// git.SetLogger(git.NewLogger(git.DebugLevel))

	path := keys.RandTempPath()
	t.Logf("Path: %s", path)
	path2 := keys.RandTempPath()

	privateKey, err := ioutil.ReadFile("id_ed25519")
	require.NoError(t, err)
	sshKey, err := keys.ParseSSHKey(privateKey, nil, true)
	require.NoError(t, err)
	repoKey, ok := sshKey.(*keys.EdX25519Key)
	require.True(t, ok)

	url := "git@gitlab.com:gabrielha/keys.pub.test.git"

	krDir := "GitTest-" + keys.Rand3262()

	salt := bytes.Repeat([]byte{0x01}, 16)
	pkey, err := keyring.KeyForPassword("testpassword", salt)
	require.NoError(t, err)
	provision := keyring.NewProvision(keyring.UnknownAuth)

	// Keyring #1
	repo1, err := git.NewRepository(git.Key(repoKey), git.KeyringDir(krDir))
	require.NoError(t, err)
	err = repo1.Clone(url, path)
	require.NoError(t, err)
	kr1, err := keyring.New(keyring.WithStore(repo1))
	require.NoError(t, err)
	err = kr1.Setup(pkey, provision)
	require.NoError(t, err)
	err = repo1.Push()
	require.NoError(t, err)

	// Keyring #2
	repo2, err := git.NewRepository(git.Key(repoKey), git.KeyringDir(krDir))
	require.NoError(t, err)
	err = repo2.Clone(url, path2)
	require.NoError(t, err)
	kr2, err := keyring.New(keyring.WithStore(repo2))
	require.NoError(t, err)
	_, err = kr2.Unlock(keys.Rand32())
	require.EqualError(t, err, "invalid keyring auth")
	_, err = kr2.Unlock(pkey)
	require.NoError(t, err)

	// Repo1: Create, Push
	item := keyring.NewItem(keys.Rand3262(), []byte("mypassword"), "", time.Now())
	err = kr1.Create(item)
	require.NoError(t, err)
	err = repo1.Push()
	require.NoError(t, err)

	// Keyring #3 (same dir as #1)
	repo3, err := git.NewRepository(git.Key(repoKey), git.KeyringDir(krDir))
	require.NoError(t, err)
	err = repo3.Open(path)
	require.NoError(t, err)
	kr3, err := keyring.New(keyring.WithStore(repo3))
	require.NoError(t, err)
	_, err = kr3.Unlock(pkey)
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
	ok, err = kr2.Delete(item.ID)
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
	git.SetLogger(git.NewLogger(git.DebugLevel))

	path := keys.RandTempPath()
	path2 := keys.RandTempPath()

	privateKey, err := ioutil.ReadFile("id_ed25519")
	require.NoError(t, err)
	sshKey, err := keys.ParseSSHKey(privateKey, nil, true)
	require.NoError(t, err)
	repoKey, ok := sshKey.(*keys.EdX25519Key)
	require.True(t, ok)

	url := "git@gitlab.com:gabrielha/keys.pub.test.git"

	krDir := "GitTest-" + keys.Rand3262()

	salt := bytes.Repeat([]byte{0x01}, 16)
	pkey, err := keyring.KeyForPassword("testpassword", salt)
	require.NoError(t, err)
	provision := keyring.NewProvision(keyring.UnknownAuth)

	// Keyring #1
	repo1, err := git.NewRepository(git.Key(repoKey), git.KeyringDir(krDir))
	require.NoError(t, err)
	err = repo1.Clone(url, path)
	require.NoError(t, err)
	kr1, err := keyring.New(keyring.WithStore(repo1))
	require.NoError(t, err)
	err = kr1.Setup(pkey, provision)
	require.NoError(t, err)
	err = repo1.Push()
	require.NoError(t, err)

	// Keyring #2
	repo2, err := git.NewRepository(git.Key(repoKey), git.KeyringDir(krDir))
	require.NoError(t, err)
	err = repo2.Clone(url, path2)
	require.NoError(t, err)
	kr2, err := keyring.New(keyring.WithStore(repo2))
	require.NoError(t, err)
	_, err = kr2.Unlock(pkey)
	require.NoError(t, err)

	// Repo1: Create, Push
	item := keyring.NewItem(keys.Rand3262(), []byte("testpassword"), "", time.Now())
	t.Logf("Create repo1 item: %s", item.ID)
	err = kr1.Create(item)
	require.NoError(t, err)
	err = repo1.Push()
	require.NoError(t, err)

	// Repo2: Create, Push (conflict)
	item.Data = []byte("testpassword2")
	t.Logf("Create repo2 item: %s", item.ID)
	err = kr2.Create(item)
	require.NoError(t, err)
	err = repo2.Push()
	require.EqualError(t, err, "non-fast-forward update: refs/heads/master")

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

	// Repo1: Edit, Push
	err = kr1.Update(item.ID, []byte("testpassword4"))
	require.NoError(t, err)
	err = repo1.Push()
	require.NoError(t, err)
	// Repo2: Edit, Pull
	err = kr2.Update(item.ID, []byte("testpassword3"))
	err = repo2.Pull()
	require.NoError(t, err)

	out, err = kr2.Get(item.ID)
	require.NoError(t, err)
	require.Equal(t, "testpassword3", string(out.Data))
}

func TestRepositoryClone(t *testing.T) {
	var err error
	// git.SetLogger(git.NewLogger(git.DebugLevel))

	urs := "git@gitlab.com:gabrielha/empty.git"
	path := keys.RandTempPath()
	// defer func() { _ = os.RemoveAll(path) }()

	privateKey, err := ioutil.ReadFile("id_ed25519")
	require.NoError(t, err)
	sshKey, err := keys.ParseSSHKey(privateKey, nil, true)
	require.NoError(t, err)
	repoKey, ok := sshKey.(*keys.EdX25519Key)
	require.True(t, ok)

	repo, err := git.NewRepository(git.Key(repoKey))
	require.NoError(t, err)
	err = repo.Clone(urs, path)
	require.NoError(t, err)

	repo2, err := git.NewRepository(git.Key(repoKey))
	require.NoError(t, err)
	err = repo2.Open(path)
	require.NoError(t, err)
}
