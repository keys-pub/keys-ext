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

	aliceClient := testClient(t, env)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	bobClient := testClient(t, env)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	// Create invite
	resp, err := aliceClient.InviteCreate(context.TODO(), alice, bob.ID())
	require.NoError(t, err)

	// Get invite
	inviteDetails, err := bobClient.Invite(context.TODO(), bob, resp.Code)
	require.NoError(t, err)
	require.Equal(t, bob.ID(), inviteDetails.Recipient)
	require.Equal(t, alice.ID(), inviteDetails.Sender)
}
