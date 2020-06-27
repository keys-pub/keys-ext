package vault_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/encoding"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	var err error

	// Vault #1 (mem)
	st := vault.NewMem()
	vlt := vault.New(st)
	require.NoError(t, err)
	key := keys.Rand32()
	id := encoding.MustEncode(bytes.Repeat([]byte{0x01}, 32), encoding.Base62)
	provision := &vault.Provision{
		ID: id,
	}
	err = vlt.Setup(key, provision)
	require.NoError(t, err)

	item := vault.NewItem(encoding.MustEncode(bytes.Repeat([]byte{0x02}, 32), encoding.Base62), []byte("testpassword"), "", time.Now())
	err = vlt.Set(item)
	require.NoError(t, err)

	// Vault #2 (mem)
	st2 := vault.NewMem()
	vlt2 := vault.New(st2)

	// Copy
	expected := []string{
		"/auth/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/db/increment",
		"/item/0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"/provision/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/push/000000000000001/auth/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/push/000000000000002/provision/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/push/000000000000003/item/0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
	}
	paths, err := vault.Copy(st, st2)
	require.NoError(t, err)
	require.Equal(t, expected, paths)

	// Unlock #2
	_, err = vlt2.Unlock(key)
	require.NoError(t, err)

	out, err := vlt2.Get(item.ID)
	require.NoError(t, err)
	require.Equal(t, "testpassword", string(out.Data))

	// Copy (again)
	_, err = vault.Copy(st, st2)
	require.EqualError(t, err, "failed to copy: already exists /auth/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29")

	// Copy (skip existing)
	paths, err = vault.Copy(st, st2, vault.SkipExisting())
	require.NoError(t, err)
	require.Equal(t, []string{}, paths)

	// Copy (dry-run)
	st3 := vault.NewMem()
	require.NoError(t, err)
	paths, err = vault.Copy(st, st3, vault.DryRun())
	require.NoError(t, err)
	require.Equal(t, expected, paths)

	docs3, err := st3.Documents()
	require.NoError(t, err)
	require.Equal(t, 0, len(docs3))
}
