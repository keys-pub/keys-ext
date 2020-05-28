package client_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys-ext/http/client"
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

	// SendMessage #1
	b1 := []byte("hi alice")
	resp, err := aliceClient.SendMessage(context.TODO(), alice, bob.ID(), b1, time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ID)

	// SendMessage #2
	b2 := []byte("what time we meeting?")
	resp, err = bobClient.SendMessage(context.TODO(), bob, alice.ID(), b2, time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ID)

	// Messages #1
	msgs, version, err := aliceClient.Messages(context.TODO(), alice, bob.ID(), nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(msgs))
	data1, pk1, err := aliceClient.DecryptMessage(alice, msgs[0])
	require.NoError(t, err)
	require.Equal(t, b1, data1)
	require.Equal(t, alice.ID(), pk1)
	data2, pk2, err := aliceClient.DecryptMessage(alice, msgs[1])
	require.NoError(t, err)
	require.Equal(t, b2, data2)
	require.Equal(t, bob.ID(), pk2)
	require.Equal(t, int64(1234567890004), tsutil.Millis(msgs[0].CreatedAt))

	// SendMessage #3
	b3 := []byte("3pm")
	resp, err = aliceClient.SendMessage(context.TODO(), alice, bob.ID(), b3, time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ID)

	// Messages #2 (from version)
	msgs, _, err = aliceClient.Messages(context.TODO(), alice, bob.ID(), &client.MessagesOpts{Version: version})
	require.NoError(t, err)
	require.Equal(t, 2, len(msgs))
	data2, pk2, err = aliceClient.DecryptMessage(alice, msgs[0])
	require.NoError(t, err)
	require.Equal(t, b2, data2)
	require.Equal(t, bob.ID(), pk2)
	data3, pk3, err := aliceClient.DecryptMessage(alice, msgs[1])
	require.NoError(t, err)
	require.Equal(t, b3, data3)
	require.Equal(t, alice.ID(), pk3)

	// Messages (desc)
	msgs, _, err = aliceClient.Messages(context.TODO(), alice, bob.ID(), &client.MessagesOpts{Direction: ds.Descending})
	require.NoError(t, err)
	require.Equal(t, 3, len(msgs))
	data1, _, err = aliceClient.DecryptMessage(alice, msgs[0])
	require.NoError(t, err)
	require.Equal(t, b3, data1)
	data2, _, err = aliceClient.DecryptMessage(alice, msgs[1])
	require.NoError(t, err)
	require.Equal(t, b2, data2)
	data3, _, err = aliceClient.DecryptMessage(alice, msgs[2])
	require.NoError(t, err)
	require.Equal(t, b1, data3)

	// Messages not found
	unknown := keys.GenerateEdX25519Key()
	msgs, _, err = aliceClient.Messages(context.TODO(), alice, unknown.ID(), nil)
	require.NoError(t, err)
	require.Empty(t, msgs)

	// Same sender/recipient
	resp, err = aliceClient.SendMessage(context.TODO(), alice, alice.ID(), []byte("selfie"), time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ID)
}
