package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestChannel(t *testing.T) {
	env, closeFn := newEnv(t)
	defer closeFn()
	testChannel(t, env, testKeysSeeded())
}

func TestChannelFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	env, closeFn := newEnvWithOptions(t, &envOptions{fi: testFirestore(t), clock: tsutil.NewTestClock()})
	defer closeFn()

	env.logger = client.NewLogger(client.DebugLevel)

	testChannel(t, env, testKeysRandom())
}

func testChannel(t *testing.T, env *env, tk testKeys) {
	alice, bob, channel := tk.alice, tk.bob, tk.channel

	aliceClient := newTestClient(t, env)
	// bobClient := newTestClient(t, env)

	err := aliceClient.ChannelCreate(context.TODO(), channel, alice)
	require.NoError(t, err)

	err = aliceClient.InviteToChannel(context.TODO(), channel, alice, bob.ID())
	require.NoError(t, err)

	invites, err := aliceClient.ChannelInvites(context.TODO(), channel, alice)
	require.NoError(t, err)
	require.Equal(t, 1, len(invites))
	require.Equal(t, bob.ID(), invites[0].Recipient)
	require.Equal(t, alice.ID(), invites[0].Sender)
	require.Equal(t, channel.ID(), invites[0].Channel)
}
