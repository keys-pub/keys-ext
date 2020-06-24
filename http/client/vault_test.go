package client_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
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

	changes := []*client.VaultChange{
		&client.VaultChange{Path: "/col1/key1", Data: []byte("test1"), Nonce: api.GenerateNonce()},
		&client.VaultChange{Path: "/col1/key2", Data: []byte("test2"), Nonce: api.GenerateNonce()},
	}
	err = aliceClient.VaultChanged(context.TODO(), alice, changes)
	require.NoError(t, err)

	vault, err := aliceClient.Vault(context.TODO(), alice)
	require.NoError(t, err)
	require.Equal(t, 2, len(vault.Changes))
	require.Equal(t, []byte("test1"), vault.Changes[0].Data)
	require.Equal(t, []byte("test2"), vault.Changes[1].Data)

	changes2 := []*client.VaultChange{
		&client.VaultChange{Path: "/col1/key3", Data: []byte("test3"), Nonce: api.GenerateNonce()},
		&client.VaultChange{Path: "/col1/key4", Data: []byte("test4.1"), Nonce: api.GenerateNonce()},
		&client.VaultChange{Path: "/col1/key4", Data: []byte("test4.2"), Nonce: api.GenerateNonce()},
		&client.VaultChange{Path: "/col1/key5", Data: []byte("test5"), Nonce: api.GenerateNonce()},
	}
	err = aliceClient.VaultChanged(context.TODO(), alice, changes2)
	require.NoError(t, err)

	vault, err = aliceClient.Vault(context.TODO(), alice, client.VaultVersion(vault.Version))
	require.NoError(t, err)
	require.Equal(t, 4, len(vault.Changes))
	require.Equal(t, "/col1/key3", vault.Changes[0].Path)
	require.Equal(t, []byte("test3"), vault.Changes[0].Data)
	require.Equal(t, "/col1/key4", vault.Changes[1].Path)
	require.Equal(t, []byte("test4.1"), vault.Changes[1].Data)
	require.Equal(t, "/col1/key4", vault.Changes[2].Path)
	require.Equal(t, []byte("test4.2"), vault.Changes[2].Data)
	require.Equal(t, "/col1/key5", vault.Changes[3].Path)
	require.Equal(t, []byte("test5"), vault.Changes[3].Data)
}
