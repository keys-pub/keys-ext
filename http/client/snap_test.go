package client

import (
	"bytes"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestSnap(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t)
	defer env.closeFn()

	ks := keys.NewMemKeystore()
	client := testClient(t, env, ks)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	err := ks.SaveEdX25519Key(alice)
	require.NoError(t, err)

	// Snap (not found)
	out, err := client.Snap(alice)
	require.NoError(t, err)
	require.Nil(t, out)

	// PutSnap
	b := []byte("secret")
	err = client.PutSnap(alice, b)
	require.NoError(t, err)

	// Snap
	out, err = client.Snap(alice)
	require.NoError(t, err)
	require.Equal(t, b, out)

	// PutSnap #2
	b2 := []byte("secret2")
	err = client.PutSnap(alice, b2)
	require.NoError(t, err)

	// Snap
	out, err = client.Snap(alice)
	require.NoError(t, err)
	require.Equal(t, b2, out)

	// Delete
	err = client.DeleteSnap(alice)
	require.NoError(t, err)

	// Snap (not found)
	out, err = client.Snap(alice)
	require.NoError(t, err)
	require.Nil(t, out)

	// Not found
	unknown := keys.GenerateEdX25519Key()
	out, err = client.Snap(unknown)
	require.NoError(t, err)
	require.Nil(t, out)
}
