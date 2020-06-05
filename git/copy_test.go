package git_test

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/git"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/keyring"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	// git.SetLogger(git.NewLogger(git.DebugLevel))
	var err error

	krDir := "GitTest-Export-" + keys.Rand3262()
	url := "git@gitlab.com:gabrielha/keys.pub.test.git"
	privateKey, err := ioutil.ReadFile("id_ed25519")
	require.NoError(t, err)
	sshKey, err := keys.ParseSSHKey(privateKey, nil, true)
	require.NoError(t, err)
	repoKey, ok := sshKey.(*keys.EdX25519Key)
	require.True(t, ok)

	// Keyring #1 (mem)
	kr := keyring.NewMem(false)
	salt, err := kr.Salt()
	require.NoError(t, err)
	key, err := keyring.KeyForPassword("testkeyringpassword", salt)
	require.NoError(t, err)
	id := encoding.MustEncode(bytes.Repeat([]byte{0x02}, 32), encoding.Base62)
	provision := &keyring.Provision{ID: id}
	err = kr.Setup(key, provision)
	require.NoError(t, err)

	iid := encoding.MustEncode(bytes.Repeat([]byte{0x03}, 32), encoding.Base62)
	item := keyring.NewItem(iid, []byte("testpassword"), "", time.Now())
	err = kr.Create(item)
	require.NoError(t, err)

	// Repo #2, Keyring #2
	path2 := keys.RandTempPath()
	repo2, err := git.NewRepository(git.Key(repoKey), git.KeyringDir(krDir))
	require.NoError(t, err)
	err = repo2.Clone(url, path2)
	require.NoError(t, err)
	kr2, err := keyring.New(keyring.WithStore(repo2))
	require.NoError(t, err)

	// Copy #1 to #2
	ids, err := keyring.Copy(kr.Store(), repo2)
	require.NoError(t, err)
	require.Equal(t, []string{
		"#auth-0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"#provision-0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"#salt",
		"0iHJbkdqdSjdOv8lotdlpRYNaigOHJYDGtSybpLpt6R",
	}, ids)

	err = repo2.Push()
	require.NoError(t, err)

	err = kr2.UnlockWithPassword("testkeyringpassword", false)
	require.NoError(t, err)

	// Repo #3, Keyring #3
	path3 := keys.RandTempPath()
	repo3, err := git.NewRepository(git.Key(repoKey), git.KeyringDir(krDir))
	require.NoError(t, err)
	err = repo3.Clone(url, path3)
	require.NoError(t, err)
	kr3, err := keyring.New(keyring.WithStore(repo3))
	require.NoError(t, err)
	err = kr3.UnlockWithPassword("testkeyringpassword", false)
	require.NoError(t, err)
	out, err := kr3.Get(item.ID)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "testpassword", string(out.Data))
}
