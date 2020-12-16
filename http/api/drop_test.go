package api_test

import (
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	kapi "github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/saltpack"
	"github.com/stretchr/testify/require"
)

func TestDrop(t *testing.T) {
	drop := api.NewChannelDrop(channel, alice.ID())
	b, err := api.Encrypt(drop, alice, bob.ID())
	require.NoError(t, err)

	out, err := api.DecryptDrop(b, saltpack.NewKeyring(bob))
	require.NoError(t, err)
	require.Equal(t, out, drop)
}

func TestDropInvalidSender(t *testing.T) {
	spoof := &api.Drop{
		Type:   api.ChannelDrop,
		Key:    kapi.NewKey(channel),
		Sender: bob.ID(),
	}
	b, err := api.Encrypt(spoof, alice, bob.ID())
	require.NoError(t, err)
	_, err = api.DecryptDrop(b, saltpack.NewKeyring(bob))
	require.EqualError(t, err, "drop sender mismatch")
}
