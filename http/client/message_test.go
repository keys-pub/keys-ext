package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

// TODO: Test truncated

func TestMessages(t *testing.T) {
	env, closeFn := newEnv(t)
	defer closeFn()

	testMessages(t, env, testKeysSeeded())
}

func TestMessagesFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	env, closeFn := newEnvWithOptions(t, &envOptions{fi: testFirestore(t), clock: tsutil.NewTestClock()})
	defer closeFn()

	testMessages(t, env, testKeysRandom())
}

func testMessages(t *testing.T, env *env, tk testKeys) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	aliceClient := newTestClient(t, env)
	bobClient := newTestClient(t, env)
	alice, bob, channel := tk.alice, tk.bob, tk.channel

	// Create channel
	err := aliceClient.ChannelCreate(context.TODO(), channel, alice)
	require.NoError(t, err)
	err = aliceClient.InviteToChannel(context.TODO(), channel, alice, bob.ID())
	require.NoError(t, err)
	err = aliceClient.ChannelInviteAccept(context.TODO(), bob, channel)
	require.NoError(t, err)

	// Messages
	msgs, err := aliceClient.Messages(context.TODO(), channel, alice, nil)
	require.NoError(t, err)
	require.Equal(t, int64(0), msgs.Index)
	require.Equal(t, 0, len(msgs.Messages))
	require.False(t, msgs.Truncated)

	// MessageSend #1
	msg1 := &api.Message{ID: "1", Text: "hi bob", Timestamp: env.clock.NowMillis()}
	err = aliceClient.MessageSend(context.TODO(), msg1, alice, channel)
	require.NoError(t, err)

	// MessageSend #2
	msg2 := &api.Message{ID: "2", Prev: "1", Text: "what time we meeting?", Timestamp: env.clock.NowMillis()}
	err = bobClient.MessageSend(context.TODO(), msg2, bob, channel)
	require.NoError(t, err)

	// Messages
	msgs, err = aliceClient.Messages(context.TODO(), channel, alice, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(msgs.Messages))
	require.False(t, msgs.Truncated)
	out1, err := client.DecryptMessage(msgs.Messages[0], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg1.Text, out1.Text)
	require.Equal(t, alice.ID(), out1.Sender)
	out2, err := client.DecryptMessage(msgs.Messages[1], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg2.Text, out2.Text)
	require.Equal(t, bob.ID(), out2.Sender)
	require.NotEmpty(t, msgs.Messages[0].Timestamp)
	require.NotEmpty(t, msgs.Messages[0].Index)

	// MessageSend #3
	msg3 := &api.Message{ID: "3", Prev: "2", Text: "3pm", Timestamp: env.clock.NowMillis()}
	err = aliceClient.MessageSend(context.TODO(), msg3, alice, channel)
	require.NoError(t, err)

	// Messages (from idx)
	msgs, err = aliceClient.Messages(context.TODO(), channel, alice, &client.MessagesOpts{Index: msgs.Index})
	require.NoError(t, err)
	require.Equal(t, 1, len(msgs.Messages))
	out3, err := client.DecryptMessage(msgs.Messages[0], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg3.Text, out3.Text)
	require.Equal(t, alice.ID(), out3.Sender)

	// Messages (desc)
	msgs, err = aliceClient.Messages(context.TODO(), channel, alice, &client.MessagesOpts{Direction: events.Descending})
	require.NoError(t, err)
	require.Equal(t, 3, len(msgs.Messages))
	out1, err = client.DecryptMessage(msgs.Messages[0], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg3.Text, out1.Text)
	out2, err = client.DecryptMessage(msgs.Messages[1], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg2.Text, out2.Text)
	out3, err = client.DecryptMessage(msgs.Messages[2], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg1.Text, out3.Text)

	// Unknown channel
	unknown := keys.GenerateEdX25519Key()
	_, err = aliceClient.Messages(context.TODO(), unknown, alice, nil)
	require.EqualError(t, err, "auth failed (403)")
}
