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
	err = aliceClient.ChannelInvite(context.TODO(), channel, alice, bob.ID())
	require.NoError(t, err)
	err = aliceClient.ChannelInviteAccept(context.TODO(), bob, channel)
	require.NoError(t, err)

	// MessageSend #1
	msg1 := &api.Message{ID: "1", Content: &api.Content{Data: []byte("hi alice"), Type: api.UTF8Content}, CreatedAt: env.clock.Now()}
	err = aliceClient.MessageSend(context.TODO(), alice, channel, msg1) // , time.Minute
	require.NoError(t, err)

	// MessageSend #2
	msg2 := &api.Message{ID: "2", Prev: "1", Content: &api.Content{Data: []byte("what time we meeting?"), Type: api.UTF8Content}, CreatedAt: env.clock.Now()}
	err = bobClient.MessageSend(context.TODO(), bob, channel, msg2) // , time.Minute
	require.NoError(t, err)

	// Messages #1
	msgs, idx, err := aliceClient.Messages(context.TODO(), channel, alice, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(msgs))
	out1, err := aliceClient.MessageDecrypt(msgs[0], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg1.Content.Data, out1.Content.Data)
	require.Equal(t, alice.ID(), out1.Sender)
	out2, err := aliceClient.MessageDecrypt(msgs[1], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg2.Content.Data, out2.Content.Data)
	require.Equal(t, bob.ID(), out2.Sender)
	require.NotEmpty(t, msgs[0].Timestamp)
	require.NotEmpty(t, msgs[0].Index)

	// MessageSend #3
	msg3 := &api.Message{ID: "3", Prev: "2", Content: &api.Content{Data: []byte("3pm"), Type: api.UTF8Content}, CreatedAt: env.clock.Now()}
	err = aliceClient.MessageSend(context.TODO(), alice, channel, msg3) // , time.Minute
	require.NoError(t, err)

	// Messages #2 (from idx)
	msgs, _, err = aliceClient.Messages(context.TODO(), channel, alice, &client.MessagesOpts{Index: idx})
	require.NoError(t, err)
	require.Equal(t, 1, len(msgs))
	out3, err := aliceClient.MessageDecrypt(msgs[0], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg3.Content.Data, out3.Content.Data)
	require.Equal(t, alice.ID(), out3.Sender)

	// Messages (desc)
	msgs, _, err = aliceClient.Messages(context.TODO(), channel, alice, &client.MessagesOpts{Direction: events.Descending})
	require.NoError(t, err)
	require.Equal(t, 3, len(msgs))
	out1, err = aliceClient.MessageDecrypt(msgs[0], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg3.Content.Data, out1.Content.Data)
	out2, err = aliceClient.MessageDecrypt(msgs[1], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg2.Content.Data, out2.Content.Data)
	out3, err = aliceClient.MessageDecrypt(msgs[2], saltpack.NewKeyring(channel))
	require.NoError(t, err)
	require.Equal(t, msg1.Content.Data, out3.Content.Data)

	// Unknown channel
	unknown := keys.GenerateEdX25519Key()
	_, _, err = aliceClient.Messages(context.TODO(), unknown, alice, nil)
	require.EqualError(t, err, "auth failed (403)")
}
