package vault_test

import (
	"sync"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/stretchr/testify/require"
)

func TestSaveKeyDelete(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	sk := keys.GenerateEdX25519Key()
	vk := vault.NewKey(sk, time.Now())
	require.NoError(t, err)
	out, updated, err := vlt.SaveKey(vk)
	require.NoError(t, err)
	require.False(t, updated)
	require.NotEmpty(t, out.CreatedAt)
	require.NotEmpty(t, out.UpdatedAt)
	key, err := vlt.Key(sk.ID())
	require.NoError(t, err)
	require.NotNil(t, key)
	skOut, err := key.AsEdX25519()
	require.NoError(t, err)
	require.Equal(t, sk.PrivateKey(), skOut.PrivateKey())
	require.Equal(t, sk.PublicKey().Bytes(), skOut.PublicKey().Bytes())

	ok, err := vlt.Delete(sk.ID().String())
	require.NoError(t, err)
	require.True(t, ok)

	out, err = vlt.Key(sk.ID())
	require.NoError(t, err)
	require.Nil(t, out)

	ok, err = vlt.Delete(sk.ID().String())
	require.NoError(t, err)
	require.False(t, ok)
}

func TestStoreConcurrent(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	sk := keys.GenerateEdX25519Key()
	key := vault.NewKey(sk, vlt.Now())
	_, _, err = vlt.SaveKey(key)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 2000; i++ {
			item, err := vlt.Key(sk.ID())
			require.NoError(t, err)
			require.NotNil(t, item)
		}
		wg.Done()
	}()
	for i := 0; i < 2000; i++ {
		item, err := vlt.Key(sk.ID())
		require.NoError(t, err)
		require.NotNil(t, item)
	}
	wg.Wait()
}

func TestExportImportKey(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	sk := keys.GenerateEdX25519Key()
	key := vault.NewKey(sk, vlt.Now())
	_, _, err = vlt.SaveKey(key)
	require.NoError(t, err)

	password := "testpassword"
	msg, err := vlt.ExportSaltpack(sk.ID(), password)
	require.NoError(t, err)

	vlt2, closeFn2 := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn2()

	out, err := vlt2.ImportSaltpack(msg, "testpassword", false)
	require.NoError(t, err)
	require.Equal(t, sk.ID(), out.ID)
}
