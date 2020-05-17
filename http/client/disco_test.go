package client_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/client"
	"github.com/stretchr/testify/require"
)

func TestDisco(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	// logger = NewLogger(DebugLevel)

	env := testEnv(t, nil)
	defer env.closeFn()

	aliceClient := testClient(t, env)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	bobClient := testClient(t, env)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	// Put
	err := aliceClient.PutDisco(context.TODO(), alice, bob.ID(), client.Offer, "hi", time.Minute)
	require.NoError(t, err)

	// Get
	out, err := bobClient.GetDisco(context.TODO(), alice.ID(), bob, client.Offer)
	require.NoError(t, err)
	require.Equal(t, "hi", out)

	// Get (again)
	out, err = bobClient.GetDisco(context.TODO(), alice.ID(), bob, client.Offer)
	require.NoError(t, err)
	require.Empty(t, out)

	// Put
	err = aliceClient.PutDisco(context.TODO(), alice, bob.ID(), client.Offer, "hi2", time.Minute)
	require.NoError(t, err)

	// Delete
	err = aliceClient.DeleteDisco(context.TODO(), alice, bob.ID())
	require.NoError(t, err)

	// Get (deleted)
	out, err = bobClient.GetDisco(context.TODO(), alice.ID(), bob, client.Offer)
	require.NoError(t, err)
	require.Empty(t, out)

	// Put
	err = aliceClient.PutDisco(context.TODO(), alice, bob.ID(), client.Offer, "hi3", time.Millisecond)
	require.NoError(t, err)

	// Get (expired)
	time.Sleep(time.Millisecond)
	out, err = bobClient.GetDisco(context.TODO(), alice.ID(), bob, client.Offer)
	require.NoError(t, err)
	require.Empty(t, out)
}
