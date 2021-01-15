package keyring_test

import (
	"sync"
	"testing"

	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys-ext/vault/keyring"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
	"github.com/stretchr/testify/require"
)

func TestSaveKeyDelete(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	kr := keyring.New(vlt)

	sk := keys.GenerateEdX25519Key()
	vk := api.NewKey(sk)
	require.NoError(t, err)

	err = kr.Save(vk)
	require.NoError(t, err)

	key, err := kr.Get(sk.ID())
	require.NoError(t, err)
	require.NotNil(t, key)
	skOut := key.AsEdX25519()
	require.NotNil(t, skOut)
	require.Equal(t, sk, skOut)

	ok, err := kr.Delete(sk.ID().String())
	require.NoError(t, err)
	require.True(t, ok)

	out, err := kr.Get(sk.ID())
	require.NoError(t, err)
	require.Nil(t, out)

	ok, err = kr.Delete(sk.ID().String())
	require.NoError(t, err)
	require.False(t, ok)
}

func TestStoreConcurrent(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()
	kr := keyring.New(vlt)

	sk := keys.GenerateEdX25519Key()
	key := api.NewKey(sk)
	err = kr.Save(key)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 2000; i++ {
			item, err := kr.Get(sk.ID())
			require.NoError(t, err)
			require.NotNil(t, item)
		}
		wg.Done()
	}()
	for i := 0; i < 2000; i++ {
		item, err := kr.Get(sk.ID())
		require.NoError(t, err)
		require.NotNil(t, item)
	}
	wg.Wait()
}

func TestExportImportKey(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()
	kr := keyring.New(vlt)

	sk := keys.GenerateEdX25519Key()
	key := api.NewKey(sk)
	err = kr.Save(key)
	require.NoError(t, err)

	password := "testpassword"
	msg, err := kr.ExportKey(sk.ID(), password)
	require.NoError(t, err)

	vault2, closeFn2 := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn2()
	kr2 := keyring.New(vault2)

	out, err := kr2.ImportKey(msg, "testpassword")
	require.NoError(t, err)
	require.Equal(t, sk.ID(), out.ID)
}

func TestKeysV1(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()
	kr := keyring.New(vlt)

	sk := keys.GenerateEdX25519Key()

	// Set v1 key
	item := vault.NewItem(sk.ID().String(), sk.Private(), "edx25519", vlt.Now())
	err = vlt.Set(item)
	require.NoError(t, err)

	// Overwrite key
	err = kr.Save(api.NewKey(sk))
	require.NoError(t, err)

	out, err := kr.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(out))
	require.Equal(t, sk.ID(), out[0].ID)
}
