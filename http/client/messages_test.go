package client

import (
	"bytes"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t)
	defer env.closeFn()

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	group := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x04}, 32)))

	ks := keys.NewKeystore()
	sp := saltpack.NewSaltpack(ks)

	// PutMessage #1
	mid1 := keys.RandString(32)
	b1 := []byte("hi alice")
	data1, err := sp.Signcrypt(b1, alice, group.ID())
	require.NoError(t, err)
	err = env.client.PutMessage(group, mid1, data1)
	require.NoError(t, err)

	// PutMessage #2
	mid2 := keys.RandString(32)
	b2 := []byte("what time we meeting?")
	data2, err := sp.Signcrypt(b2, bob, group.ID())
	require.NoError(t, err)
	err = env.client.PutMessage(group, mid2, data2)
	require.NoError(t, err)

	// Messages #1
	respA1, err := env.client.Messages(group, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(respA1.Messages))
	require.Equal(t, mid1, respA1.Messages[0].ID)
	require.Equal(t, mid2, respA1.Messages[1].ID)
	ts0 := keys.TimeToMillis(respA1.MetadataFor(respA1.Messages[0]).CreatedAt)
	require.Equal(t, keys.TimeMs(1234567890004), ts0)
	require.Equal(t, data1, respA1.Messages[0].Data)
	require.Equal(t, data2, respA1.Messages[1].Data)

	// PutMessage #3
	mid3 := keys.RandString(32)
	b3 := []byte("3pm")
	data3, err := sp.Signcrypt(b3, alice, group.ID())
	require.NoError(t, err)
	err = env.client.PutMessage(group, mid3, data3)
	require.NoError(t, err)

	// Messages #2 (from version)
	respA2, errA2 := env.client.Messages(group, respA1.Version)
	require.NoError(t, errA2)
	require.Equal(t, 2, len(respA2.Messages))
	require.Equal(t, mid2, respA2.Messages[0].ID)
	require.Equal(t, mid3, respA2.Messages[1].ID)
	require.Equal(t, data3, respA2.Messages[1].Data)

	// Messages not found
	unknown := keys.GenerateEdX25519Key()
	resp, err := env.client.Messages(unknown, "")
	require.NoError(t, err)
	require.Nil(t, resp)
}
