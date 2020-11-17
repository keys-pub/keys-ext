package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestInbox(t *testing.T) {
	env, closeFn := newEnv(t)
	defer closeFn()
	testInbox(t, env, testKeysSeeded())
}

func TestInboxFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	env, closeFn := newEnvWithOptions(t, &envOptions{fi: testFirestore(t), clock: tsutil.NewTestClock()})
	defer closeFn()

	testInbox(t, env, testKeysRandom())
}

func testInbox(t *testing.T, env *env, tk testKeys) {
	aliceClient := newTestClient(t, env)
	// bobClient := newTestClient(t, env)

	alice, channel := tk.alice, tk.channel

	err := aliceClient.ChannelCreate(context.TODO(), channel, alice)
	require.NoError(t, err)

	// Channels
	channels, err := aliceClient.InboxChannels(context.TODO(), alice)
	require.NoError(t, err)
	require.Equal(t, 1, len(channels))
	require.Equal(t, channel.ID(), channels[0].ID)
	require.Equal(t, int64(0), channels[0].Index)
	// require.Equal(t, int64(0), channels[0].Timestamp)

	// MessageSend #1
	msg1 := &api.Message{ID: "1", Text: "hi bob", Timestamp: env.clock.NowMillis()}
	err = aliceClient.MessageSend(context.TODO(), msg1, alice, channel)
	require.NoError(t, err)

	channels, err = aliceClient.InboxChannels(context.TODO(), alice)
	require.NoError(t, err)
	require.Equal(t, 1, len(channels))
	require.Equal(t, channel.ID(), channels[0].ID)
	require.Equal(t, int64(1), channels[0].Index)
	// require.Equal(t, int64(1234567890016), channels[0].Timestamp)
}
