package client

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestEphem(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t, logger)
	defer env.closeFn()

	ksa := keys.NewMemKeystore()
	aliceClient := testClient(t, env, ksa)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	err := ksa.SaveEdX25519Key(alice)
	require.NoError(t, err)

	ksb := keys.NewMemKeystore()
	bobClient := testClient(t, env, ksb)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	err = ksb.SaveEdX25519Key(bob)
	require.NoError(t, err)

	// Put
	err = aliceClient.PutEphemeral(context.TODO(), alice.ID(), bob.ID(), "offer", []byte("hi"))
	require.NoError(t, err)

	// Get
	out, err := bobClient.GetEphemeral(context.TODO(), bob.ID(), alice.ID(), "offer")
	require.NoError(t, err)
	require.Equal(t, []byte("hi"), out)

	// Get (again)
	out, err = bobClient.GetEphemeral(context.TODO(), bob.ID(), alice.ID(), "offer")
	require.NoError(t, err)
	require.Nil(t, out)
}
