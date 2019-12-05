package client

import (
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestItems(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	clock := newClock()
	testClient := testClient(t, clock)
	cl := testClient.client
	defer testClient.closeFn()

	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)
	bob, err := keys.NewKeyFromSeedPhrase(bobSeed, false)
	require.NoError(t, err)
	group, err := keys.NewKeyFromSeedPhrase(groupSeed, false)
	require.NoError(t, err)

	errA1 := testClient.ks.SaveKey(alice, true, clock.Now())
	require.NoError(t, errA1)
	errB2 := testClient.ks.SaveKey(bob, true, clock.Now())
	require.NoError(t, errB2)
	errG1 := testClient.ks.SaveKey(group, true, clock.Now())
	require.NoError(t, errG1)

	// PutItem #1
	mid1 := keys.RandID()
	_, err = cl.PutItem(bob, group, mid1, []byte("password1"))
	require.NoError(t, err)

	// PutItem #2
	mid2 := keys.RandID()
	_, err = cl.PutItem(alice, group, mid2, []byte("password2"))
	require.NoError(t, err)

	// Vault #1
	respA1, errA1 := cl.Vault(group, "")
	require.NoError(t, errA1)
	require.Equal(t, 2, len(respA1.Items))
	require.Equal(t, mid1, respA1.Items[0].ID)
	require.Equal(t, mid2, respA1.Items[1].ID)
	ts0 := keys.TimeToMillis(respA1.MetadataFor(respA1.Items[0]).CreatedAt)
	require.Equal(t, keys.TimeMs(1234567890007), ts0)
	out, sender, err := cl.cp.Open(respA1.Items[0].Data)
	require.NoError(t, err)
	require.Equal(t, bob.ID(), sender)
	require.Equal(t, "password1", string(out))

	// PutItem #3
	mid3 := keys.RandID()
	_, err = cl.PutItem(bob, group, mid3, []byte("password3"))
	require.NoError(t, err)

	// Vault #2 (from version)
	respA2, errA2 := cl.Vault(group, respA1.Version)
	require.NoError(t, errA2)
	require.Equal(t, 2, len(respA2.Items))
	require.Equal(t, mid2, respA2.Items[0].ID)
	require.Equal(t, mid3, respA2.Items[1].ID)
	out2, sender2, err := cl.cp.Open(respA2.Items[1].Data)
	require.NoError(t, err)
	require.Equal(t, bob.ID(), sender2)
	require.Equal(t, "password3", string(out2))

	// Vault not found
	key := keys.GenerateKey()
	resp, err := cl.Vault(key, "")
	require.NoError(t, err)
	require.Nil(t, resp)
}
