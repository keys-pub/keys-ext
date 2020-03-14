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

	// PostMessage #1
	b1 := []byte("hi alice")
	msg, err := client.PostMessage(alice, alice.ID(), b1)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	// PutMessage #2
	b2 := []byte("what time we meeting?")
	msg, err = client.PostMessage(alice, alice.ID(), b2)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	// Messages #1
	resp, err := client.Messages(alice, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Messages))
	require.Equal(t, b1, resp.Messages[0].Data)
	require.Equal(t, b2, resp.Messages[1].Data)
	ts0 := keys.TimeToMillis(resp.MetadataFor(resp.Messages[0]).CreatedAt)
	require.Equal(t, keys.TimeMs(1234567890004), ts0)

	// PostMessage #3
	b3 := []byte("3pm")
	msg, err = client.PostMessage(alice, alice.ID(), b3)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	// Messages #2 (from version)
	resp, err = client.Messages(alice, resp.Version)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Messages))
	require.Equal(t, b2, resp.Messages[0].Data)
	require.Equal(t, b3, resp.Messages[1].Data)

	// Messages not found
	unknown := keys.GenerateEdX25519Key()
	resp, err = client.Messages(unknown, "")
	require.NoError(t, err)
	require.Nil(t, resp)
}
