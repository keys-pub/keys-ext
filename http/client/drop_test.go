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

func TestDrop(t *testing.T) {
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

	// Drop auth
	bobToken := "bobtoken"
	err := bobClient.DropAuth(ctx, bob, bobToken)
	require.NoError(t, err)

	// Drops
	msgs, err := aliceClient.Drops(ctx, alice, nil)
	require.NoError(t, err)
	require.Equal(t, int64(0), msgs.Index)
	require.Equal(t, 0, len(msgs.Events))
	require.False(t, msgs.Truncated)

	// Drop #1
	msg1 := api.NewMessage(alice.ID()).WithText("hi bob").WithTimestamp(env.clock.NowMillis())
	err = aliceClient.Drop(ctx, msg1, alice, bob.ID(), bobToken)
	require.NoError(t, err)

	var out1 *api.Message
	var out3 *api.Message

	// Drops
	msgs, err = bobClient.Drops(ctx, bob, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(msgs.Events))
	require.False(t, msgs.Truncated)
	out1, err = api.DecryptMessageFromEvent(msgs.Events[0], bob)
	require.NoError(t, err)
	require.Equal(t, msg1.Text, out1.Text)
	require.Equal(t, alice.ID(), out1.Sender)

	// Drop #3
	msg3 := api.NewMessage(alice.ID()).WithText("here it is").WithTimestamp(env.clock.NowMillis())
	err = aliceClient.Drop(ctx, msg3, alice, bob.ID(), bobToken)
	require.NoError(t, err)

	// Drops (from idx)
	msgs, err = bobClient.Drops(ctx, bob, &client.MessagesOpts{Index: msgs.Index})
	require.NoError(t, err)
	require.Equal(t, 1, len(msgs.Events))
	out3, err = api.DecryptMessageFromEvent(msgs.Events[0], bob)
	require.NoError(t, err)
	require.Equal(t, msg3.Text, out3.Text)
	require.Equal(t, alice.ID(), out3.Sender)

	// Drops (desc)
	msgs, err = bobClient.Drops(ctx, bob, &client.MessagesOpts{Order: events.Descending})
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
	msgs, err = aliceClient.Drops(ctx, unknown, nil)
	require.NoError(t, err)
	require.Empty(t, msgs.Events)
}
