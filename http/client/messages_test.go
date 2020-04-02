package client

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t, logger)
	defer env.closeFn()

	ksa := keys.NewMemKeystore()
	aliceClient := testClient(t, env, ksa)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	err := ksa.SaveEdX25519Key(alice)
	require.NoError(t, err)

	ksb := keys.NewMemKeystore()
	bobClient := testClient(t, env, ksb)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	err = ksb.SaveEdX25519Key(bob)
	require.NoError(t, err)

	// SendMessage #1
	id1 := keys.Rand3262()
	b1 := []byte("hi alice")
	err = aliceClient.SendMessage(context.TODO(), alice.ID(), bob.ID(), "default", id1, b1, time.Minute)
	require.NoError(t, err)

	// SendMessage #2
	id2 := keys.Rand3262()
	b2 := []byte("what time we meeting?")
	err = bobClient.SendMessage(context.TODO(), bob.ID(), alice.ID(), "default", id2, b2, time.Minute)
	require.NoError(t, err)

	// Messages #1
	msgs, version, err := aliceClient.Messages(context.TODO(), alice.ID(), bob.ID(), "default", nil)
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
	require.Equal(t, keys.TimeMs(1234567890004), keys.TimeToMillis(msgs[0].CreatedAt))

	// SendMessage #3
	id3 := keys.Rand3262()
	b3 := []byte("3pm")
	err = aliceClient.SendMessage(context.TODO(), alice.ID(), bob.ID(), "default", id3, b3, time.Minute)
	require.NoError(t, err)

	// Messages #2 (from version)
	msgs, _, err = aliceClient.Messages(context.TODO(), alice.ID(), bob.ID(), "default", &MessagesOpts{Version: version})
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
	msgs, _, err = aliceClient.Messages(context.TODO(), alice.ID(), bob.ID(), "default", &MessagesOpts{Direction: keys.Descending})
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
	msgs, _, err = aliceClient.Messages(context.TODO(), alice.ID(), unknown.ID(), "default", nil)
	require.NoError(t, err)
	require.Empty(t, msgs)

	// Same sender/recipient
	err = aliceClient.SendMessage(context.TODO(), alice.ID(), alice.ID(), "default", keys.Rand3262(), []byte("selfie"), time.Minute)
	require.NoError(t, err)
}

func TestMessageExpiring(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	// logger = NewLogger(DebugLevel)

	env := testEnv(t, logger)
	defer env.closeFn()

	ksa := keys.NewMemKeystore()
	aliceClient := testClient(t, env, ksa)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	err := ksa.SaveEdX25519Key(alice)
	require.NoError(t, err)

	ksb := keys.NewMemKeystore()
	bobClient := testClient(t, env, ksb)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	err = ksb.SaveEdX25519Key(bob)
	require.NoError(t, err)

	// Put
	err = aliceClient.SendMessage(context.TODO(), alice.ID(), bob.ID(), "wormhole", "offer", []byte("hi"), time.Hour)
	require.NoError(t, err)

	// Get
	out, err := bobClient.Message(context.TODO(), bob.ID(), alice.ID(), "wormhole", "offer")
	require.NoError(t, err)
	require.Equal(t, []byte("hi"), out)

	// Put
	err = aliceClient.SendMessage(context.TODO(), alice.ID(), bob.ID(), "wormhole", "offer", []byte("hi2"), time.Hour)
	require.NoError(t, err)

	// Delete
	err = aliceClient.DeleteMessage(context.TODO(), alice.ID(), bob.ID(), "wormhole", "offer")
	require.NoError(t, err)

	// Get (deleted)
	out, err = bobClient.Message(context.TODO(), bob.ID(), alice.ID(), "wormhole", "offer")
	require.NoError(t, err)
	require.Nil(t, out)

	// Put
	err = aliceClient.SendMessage(context.TODO(), alice.ID(), bob.ID(), "wormhole", "offer", []byte("hi3"), time.Millisecond)
	require.NoError(t, err)

	// Get (expired)
	time.Sleep(time.Millisecond)
	out, err = bobClient.Message(context.TODO(), bob.ID(), alice.ID(), "wormhole", "offer")
	require.NoError(t, err)
	require.Nil(t, out)
}
