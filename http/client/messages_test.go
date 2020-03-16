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

	ks := keys.NewMemKeystore()
	client := testClient(t, env, ks)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	err := ks.SaveEdX25519Key(alice)
	require.NoError(t, err)

	// SendMessage #1
	b1 := []byte("hi alice")
	msg, err := client.SendMessage(alice, alice.ID(), b1)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	// SendMessage #2
	b2 := []byte("what time we meeting?")
	msg, err = client.SendMessage(alice, alice.ID(), b2)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	// Messages #1
	msgs, version, err := client.Messages(alice, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(msgs))
	data1, pk1, err := client.DecryptMessage(alice, msgs[0])
	require.NoError(t, err)
	require.Equal(t, b1, data1)
	require.Equal(t, alice.ID(), pk1)
	data2, pk2, err := client.DecryptMessage(alice, msgs[1])
	require.NoError(t, err)
	require.Equal(t, b2, data2)
	require.Equal(t, alice.ID(), pk2)
	require.Equal(t, keys.TimeMs(1234567890004), keys.TimeToMillis(msgs[0].CreatedAt))

	// SendMessage #3
	b3 := []byte("3pm")
	msg, err = client.SendMessage(alice, alice.ID(), b3)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	// Messages #2 (from version)
	msgs, _, err = client.Messages(alice, &MessagesOpts{Version: version})
	require.NoError(t, err)
	require.Equal(t, 2, len(msgs))
	data2, pk2, err = client.DecryptMessage(alice, msgs[0])
	require.NoError(t, err)
	require.Equal(t, b2, data2)
	require.Equal(t, alice.ID(), pk2)
	data3, pk3, err := client.DecryptMessage(alice, msgs[1])
	require.NoError(t, err)
	require.Equal(t, b3, data3)
	require.Equal(t, alice.ID(), pk3)

	// Messages (desc)
	msgs, _, err = client.Messages(alice, &MessagesOpts{Direction: keys.Descending})
	require.NoError(t, err)
	require.Equal(t, 3, len(msgs))
	data1, _, err = client.DecryptMessage(alice, msgs[0])
	require.NoError(t, err)
	require.Equal(t, b3, data1)
	data2, _, err = client.DecryptMessage(alice, msgs[1])
	require.NoError(t, err)
	require.Equal(t, b2, data2)
	data3, _, err = client.DecryptMessage(alice, msgs[2])
	require.NoError(t, err)
	require.Equal(t, b1, data3)

	// Messages not found
	unknown := keys.GenerateEdX25519Key()
	msgs, _, err = client.Messages(unknown, nil)
	require.NoError(t, err)
	require.Empty(t, msgs)
}
