package client_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/stretchr/testify/require"
)

func TestVault(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	var err error
	env := testEnv(t, nil) // client.NewLogger(client.DebugLevel)
	defer env.closeFn()

	aliceClient := testClient(t, env)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	items := []*client.VaultItem{
		&client.VaultItem{Data: []byte("test1")},
		&client.VaultItem{Data: []byte("test2")},
	}
	err = aliceClient.VaultSave(context.TODO(), alice, items)
	require.NoError(t, err)

	vault, err := aliceClient.Vault(context.TODO(), alice)
	require.NoError(t, err)
	require.Equal(t, 2, len(vault.Items))
	require.Equal(t, []byte("test1"), vault.Items[0].Data)
	require.Equal(t, []byte("test2"), vault.Items[1].Data)

	items2 := []*client.VaultItem{
		&client.VaultItem{Data: []byte("test3")},
		&client.VaultItem{Data: []byte("test4.1")},
		&client.VaultItem{Data: []byte("test4.2")},
		&client.VaultItem{Data: []byte("test5")},
	}
	err = aliceClient.VaultSave(context.TODO(), alice, items2)
	require.NoError(t, err)

	vault, err = aliceClient.Vault(context.TODO(), alice, client.VaultVersion(vault.Version))
	require.NoError(t, err)
	require.Equal(t, 5, len(vault.Items))
	require.Equal(t, []byte("test2"), vault.Items[0].Data)
	require.Equal(t, []byte("test3"), vault.Items[1].Data)
	require.Equal(t, []byte("test4.1"), vault.Items[2].Data)
	require.Equal(t, []byte("test4.2"), vault.Items[3].Data)
	require.Equal(t, []byte("test5"), vault.Items[4].Data)
}
