package client

import (
	"bytes"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t)
	defer env.closeFn()

	aliceKs := keys.NewMemKeystore()
	aliceClient := testClient(t, env, aliceKs)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	err := aliceKs.SaveEdX25519Key(alice)
	require.NoError(t, err)

	bobKs := keys.NewMemKeystore()
	bobClient := testClient(t, env, bobKs)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	err = bobKs.SaveEdX25519Key(bob)
	require.NoError(t, err)

	// SendMessage #1
	b1 := []byte("hi alice")
	msg, err := aliceClient.SendMessage(alice, bob.ID(), b1, nil)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	// SendMessage #2
	b2 := []byte("what time we meeting?")
	msg, err = bobClient.SendMessage(bob, alice.ID(), b2, nil)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	// Messages #1
	msgs, version, err := aliceClient.Messages(alice, bob.ID(), nil)
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
	b3 := []byte("3pm")
	msg, err = aliceClient.SendMessage(alice, bob.ID(), b3, nil)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	// Messages #2 (from version)
	msgs, _, err = aliceClient.Messages(alice, bob.ID(), &MessagesOpts{Version: version})
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
	msgs, _, err = aliceClient.Messages(alice, bob.ID(), &MessagesOpts{Direction: keys.Descending})
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
	msgs, _, err = aliceClient.Messages(alice, unknown.ID(), nil)
	require.NoError(t, err)
	require.Empty(t, msgs)
}
