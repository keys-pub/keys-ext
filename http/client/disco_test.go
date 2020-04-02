package client

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestDisco(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	// logger = NewLogger(DebugLevel)

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
	err = aliceClient.PutDisco(context.TODO(), alice.ID(), bob.ID(), Offer, "hi", time.Minute)
	require.NoError(t, err)

	// Get
	out, err := bobClient.GetDisco(context.TODO(), alice.ID(), bob.ID(), Offer)
	require.NoError(t, err)
	require.Equal(t, "hi", out)

	// Get (again)
	out, err = bobClient.GetDisco(context.TODO(), alice.ID(), bob.ID(), Offer)
	require.NoError(t, err)
	require.Empty(t, out)

	// Put
	err = aliceClient.PutDisco(context.TODO(), alice.ID(), bob.ID(), Offer, "hi2", time.Minute)
	require.NoError(t, err)

	// Delete
	err = aliceClient.DeleteDisco(context.TODO(), alice.ID(), bob.ID())
	require.NoError(t, err)

	// Get (deleted)
	out, err = bobClient.GetDisco(context.TODO(), alice.ID(), bob.ID(), Offer)
	require.NoError(t, err)
	require.Empty(t, out)

	// Put
	err = aliceClient.PutDisco(context.TODO(), alice.ID(), bob.ID(), Offer, "hi3", time.Millisecond)
	require.NoError(t, err)

	// Get (expired)
	time.Sleep(time.Millisecond)
	out, err = bobClient.GetDisco(context.TODO(), alice.ID(), bob.ID(), Offer)
	require.NoError(t, err)
	require.Empty(t, out)
}
