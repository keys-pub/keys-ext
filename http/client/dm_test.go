package client_test

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/stretchr/testify/require"
)

func TestDirectMessages(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env, closeFn := newEnv(t)
	defer closeFn()
	tk := testKeysSeeded()

	aliceClient := newTestClient(t, env)
	bobClient := newTestClient(t, env)
	alice, bob := tk.alice, tk.bob
	ctx := context.TODO()

	// Follow
	err := bobClient.Follow(ctx, bob, alice.ID())
	require.NoError(t, err)

	// DirectToken
	token, err := aliceClient.DirectToken(ctx, alice)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// DirectMessages
	msgs, err := aliceClient.DirectMessages(ctx, alice, nil)
	require.NoError(t, err)
	require.Equal(t, int64(0), msgs.Index)
	require.Equal(t, 0, len(msgs.Events))
	require.False(t, msgs.Truncated)

	// DirectMessageSend #1
	msg1 := api.NewMessage(alice.ID()).WithText("hi bob").WithTimestamp(env.clock.NowMillis())
	err = aliceClient.DirectMessageSend(ctx, msg1, alice, bob.ID())
	require.NoError(t, err)

	var out1 *api.Message
	var out3 *api.Message

	// DirectMessages
	msgs, err = bobClient.DirectMessages(ctx, bob, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(msgs.Events))
	require.False(t, msgs.Truncated)
	out1, err = api.DecryptMessageFromEvent(msgs.Events[0], bob)
	require.NoError(t, err)
	require.Equal(t, msg1.Text, out1.Text)
	require.Equal(t, alice.ID(), out1.Sender)

	// DirectMessageSend #3
	msg3 := api.NewMessage(alice.ID()).WithText("here it is").WithTimestamp(env.clock.NowMillis())
	err = aliceClient.DirectMessageSend(ctx, msg3, alice, bob.ID())
	require.NoError(t, err)

	// DirectMessages (from idx)
	msgs, err = bobClient.DirectMessages(ctx, bob, &client.MessagesOpts{Index: msgs.Index})
	require.NoError(t, err)
	require.Equal(t, 1, len(msgs.Events))
	out3, err = api.DecryptMessageFromEvent(msgs.Events[0], bob)
	require.NoError(t, err)
	require.Equal(t, msg3.Text, out3.Text)
	require.Equal(t, alice.ID(), out3.Sender)

	// DirectMessages (desc)
	msgs, err = bobClient.DirectMessages(ctx, bob, &client.MessagesOpts{Order: events.Descending})
	require.NoError(t, err)
	require.Equal(t, 2, len(msgs.Events))
	out1, err = api.DecryptMessageFromEvent(msgs.Events[0], bob)
	require.NoError(t, err)
	require.Equal(t, msg3.Text, out1.Text)
	out3, err = api.DecryptMessageFromEvent(msgs.Events[1], bob)
	require.NoError(t, err)
	require.Equal(t, msg1.Text, out3.Text)

	// Unknown
	unknown := keys.GenerateEdX25519Key()
	msgs, err = aliceClient.DirectMessages(ctx, unknown, nil)
	require.NoError(t, err)
	require.Empty(t, msgs.Events)
}
