package client_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestVault(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t, nil)
	defer env.closeFn()

	aliceClient := testClient(t, env)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	b1 := []byte("hi alice")
	err := aliceClient.VaultSave(context.TODO(), alice, b1)
	require.NoError(t, err)

	resp, err := aliceClient.Vault(context.TODO(), alice, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Items))
	require.Equal(t, b1, resp.Items[0].Data)
}
