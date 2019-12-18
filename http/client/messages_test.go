package client

import (
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

	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)
	bob, err := keys.NewKeyFromSeedPhrase(bobSeed, false)
	require.NoError(t, err)
	group, err := keys.NewKeyFromSeedPhrase(groupSeed, false)
	require.NoError(t, err)

	errA1 := env.ks.SaveKey(alice, true, env.clock.Now())
	require.NoError(t, errA1)
	errB2 := env.ks.SaveKey(bob, true, env.clock.Now())
	require.NoError(t, errB2)
	errG1 := env.ks.SaveKey(group, true, env.clock.Now())
	require.NoError(t, errG1)

	// PutMessage #1
	mid1 := keys.RandID()
	_, err = env.client.PutMessage(bob, group, mid1, []byte("hi alice"))
	require.NoError(t, err)

	// PutMessage #2
	mid2 := keys.RandID()
	_, err = env.client.PutMessage(alice, group, mid2, []byte("what time we meeting?"))
	require.NoError(t, err)

	// Messages #1
	respA1, errA1 := env.client.Messages(group, "")
	require.NoError(t, errA1)
	require.Equal(t, 2, len(respA1.Messages))
	require.Equal(t, mid1, respA1.Messages[0].ID)
	require.Equal(t, mid2, respA1.Messages[1].ID)
	ts0 := keys.TimeToMillis(respA1.MetadataFor(respA1.Messages[0]).CreatedAt)
	require.Equal(t, keys.TimeMs(1234567890007), ts0)
	out, sender, err := env.crypto.Open(respA1.Messages[0].Data)
	require.NoError(t, err)
	require.Equal(t, bob.ID(), sender)
	require.Equal(t, "hi alice", string(out))

	// PutMessage #3
	mid3 := keys.RandID()
	_, err = env.client.PutMessage(bob, group, mid3, []byte("3pm"))
	require.NoError(t, err)

	// Messages #2 (from version)
	respA2, errA2 := env.client.Messages(group, respA1.Version)
	require.NoError(t, errA2)
	require.Equal(t, 2, len(respA2.Messages))
	require.Equal(t, mid2, respA2.Messages[0].ID)
	require.Equal(t, mid3, respA2.Messages[1].ID)
	out2, sender2, err := env.crypto.Open(respA2.Messages[1].Data)
	require.NoError(t, err)
	require.Equal(t, bob.ID(), sender2)
	require.Equal(t, "3pm", string(out2))

	// Messages not found
	unknown := keys.GenerateKey()
	resp, err := env.client.Messages(unknown, "")
	require.NoError(t, err)
	require.Nil(t, resp)
}
