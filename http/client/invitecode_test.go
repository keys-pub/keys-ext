package client_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestInviteCode(t *testing.T) {
	// client.SetLogger(client.NewLogger(client.DebugLevel))
	// api.SetLogger(client.NewLogger(DebugLevel))

	env, closeFn := newEnvWithOptions(t, &envOptions{
		// logger: client.NewLogger(client.DebugLevel),
	})
	defer closeFn()

	aliceClient := newTestClient(t, env)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	bobClient := newTestClient(t, env)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	// Create invite
	resp, err := aliceClient.InviteCodeCreate(context.TODO(), alice, bob.ID())
	require.NoError(t, err)

	// Get invite
	invite, err := bobClient.InviteCode(context.TODO(), bob, resp.Code)
	require.NoError(t, err)
	require.NotNil(t, invite)
	require.Equal(t, bob.ID(), invite.Recipient)
	require.Equal(t, alice.ID(), invite.Sender)
}
