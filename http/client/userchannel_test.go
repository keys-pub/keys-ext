package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestUserChannel(t *testing.T) {
	env, closeFn := newEnv(t)
	defer closeFn()
	testUserChannel(t, env, testKeysSeeded())
}

func TestUserChannelFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	env, closeFn := newEnvWithOptions(t, &envOptions{fi: testFirestore(t), clock: tsutil.NewTestClock()})
	defer closeFn()

	testUserChannel(t, env, testKeysRandom())
}

func testUserChannel(t *testing.T, env *env, tk testKeys) {
	aliceClient := newTestClient(t, env)
	bobClient := newTestClient(t, env)

	alice, bob, channel := tk.alice, tk.bob, tk.channel
	ctx := context.TODO()

	info := &api.ChannelInfo{Name: "test"}
	err := aliceClient.ChannelCreate(ctx, channel, alice, info)
	require.NoError(t, err)

	// Channels
	channels, err := aliceClient.Channels(ctx, alice)
	require.NoError(t, err)
	require.Equal(t, 1, len(channels))
	require.Equal(t, channel.ID(), channels[0].ID)
	require.Equal(t, int64(1), channels[0].Index)
	// require.Equal(t, int64(0), channels[0].Timestamp)

	// MessageSend #1
	msg1 := &api.Message{ID: "1", Text: "hi bob", Timestamp: env.clock.NowMillis()}
	err = aliceClient.MessageSend(ctx, msg1, alice, channel)
	require.NoError(t, err)

	channels, err = aliceClient.Channels(ctx, alice)
	require.NoError(t, err)
	require.Equal(t, 1, len(channels))
	require.Equal(t, channel.ID(), channels[0].ID)
	require.Equal(t, int64(2), channels[0].Index)
	// require.Equal(t, int64(1234567890016), channels[0].Timestamp)

	// Invite bob
	err = aliceClient.InviteToChannel(ctx, channel, alice, bob.ID())
	require.NoError(t, err)
	// Bob join
	err = bobClient.ChannelJoin(ctx, bob, channel)
	require.NoError(t, err)

	// Leave channel
	err = aliceClient.ChannelLeave(ctx, alice, channel.ID())
	require.NoError(t, err)

	// Channels
	channels, err = aliceClient.Channels(ctx, alice)
	require.Equal(t, 0, len(channels))

	msg2 := &api.Message{ID: "2", Text: "test", Timestamp: env.clock.NowMillis()}
	err = aliceClient.MessageSend(ctx, msg2, alice, channel)
	require.EqualError(t, err, "auth failed (403)")

	// Try to re-join without invite
	err = aliceClient.ChannelJoin(ctx, alice, channel)
	require.EqualError(t, err, "invite not found (404)")
}
