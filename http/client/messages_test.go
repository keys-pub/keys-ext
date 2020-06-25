package client_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t, nil)
	defer env.closeFn()

	aliceClient := testClient(t, env)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	bobClient := testClient(t, env)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	// MessageSend #1
	msg1 := client.NewEvent("/msgs/1", []byte("hi alice"), nil)
	err := aliceClient.MessageSend(context.TODO(), alice, bob.ID(), msg1) // , time.Minute
	require.NoError(t, err)

	// MessageSend #2
	msg2 := client.NewEvent("/msgs/2", []byte("what time we meeting?"), msg1)
	err = bobClient.MessageSend(context.TODO(), bob, alice.ID(), msg2) // , time.Minute
	require.NoError(t, err)

	// Messages #1
	msgs, idx, err := aliceClient.Messages(context.TODO(), alice, bob.ID(), nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(msgs))
	out1, pk1, err := aliceClient.MessageDecrypt(alice, msgs[0])
	require.NoError(t, err)
	require.Equal(t, msg1.Data, out1.Data)
	require.Equal(t, alice.ID(), pk1)
	out2, pk2, err := aliceClient.MessageDecrypt(alice, msgs[1])
	require.NoError(t, err)
	require.Equal(t, msg2.Data, out2.Data)
	require.Equal(t, bob.ID(), pk2)
	require.Equal(t, int64(1234567890004), tsutil.Millis(msgs[0].Timestamp))

	// MessageSend #3
	msg3 := client.NewEvent("/msgs/3", []byte("3pm"), msg2)
	err = aliceClient.MessageSend(context.TODO(), alice, bob.ID(), msg3) // , time.Minute
	require.NoError(t, err)

	// Messages #2 (from idx)
	msgs, _, err = aliceClient.Messages(context.TODO(), alice, bob.ID(), &client.MessagesOpts{Index: idx})
	require.NoError(t, err)
	require.Equal(t, 1, len(msgs))
	out3, pk3, err := aliceClient.MessageDecrypt(alice, msgs[0])
	require.NoError(t, err)
	require.Equal(t, msg3.Data, out3.Data)
	require.Equal(t, alice.ID(), pk3)

	// Messages (desc)
	msgs, _, err = aliceClient.Messages(context.TODO(), alice, bob.ID(), &client.MessagesOpts{Direction: ds.Descending})
	require.NoError(t, err)
	require.Equal(t, 3, len(msgs))
	out1, _, err = aliceClient.MessageDecrypt(alice, msgs[0])
	require.NoError(t, err)
	require.Equal(t, msg3.Data, out1.Data)
	out2, _, err = aliceClient.MessageDecrypt(alice, msgs[1])
	require.NoError(t, err)
	require.Equal(t, msg2.Data, out2.Data)
	out3, _, err = aliceClient.MessageDecrypt(alice, msgs[2])
	require.NoError(t, err)
	require.Equal(t, msg1.Data, out3.Data)

	// Messages not found
	unknown := keys.GenerateEdX25519Key()
	msgs, _, err = aliceClient.Messages(context.TODO(), alice, unknown.ID(), nil)
	require.NoError(t, err)
	require.Empty(t, msgs)

	// Same sender/recipient
	self := client.NewEvent("/msgs/self", []byte("selfie"), nil)
	err = aliceClient.MessageSend(context.TODO(), alice, alice.ID(), self) // , time.Minute)
	require.NoError(t, err)
}
