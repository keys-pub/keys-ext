package client_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestInvite(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t, nil)
	defer env.closeFn()

	ksa := keys.NewMemStore(true)
	aliceClient := testClient(t, env, ksa)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	err := ksa.SaveEdX25519Key(alice)
	require.NoError(t, err)

	ksb := keys.NewMemStore(true)
	bobClient := testClient(t, env, ksb)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	err = ksb.SaveEdX25519Key(bob)
	require.NoError(t, err)

	// Create invite
	resp, err := aliceClient.CreateInvite(context.TODO(), alice.ID(), bob.ID())
	require.NoError(t, err)

	// Get invite
	inviteDetails, err := bobClient.Invite(context.TODO(), bob.ID(), resp.Code)
	require.NoError(t, err)
	require.Equal(t, bob.ID(), inviteDetails.Recipient)
	require.Equal(t, alice.ID(), inviteDetails.Sender)
}
