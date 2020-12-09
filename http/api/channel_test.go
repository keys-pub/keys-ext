package api_test

import (
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/saltpack"
	"github.com/stretchr/testify/require"
)

func TestChannelInvite(t *testing.T) {
	info := &api.ChannelInfo{Name: "test"}
	invite, err := api.NewChannelInvite(channel, info, alice, bob.ID())
	require.NoError(t, err)

	out, pk, err := invite.DecryptKey(saltpack.NewKeyring(alice))
	require.NoError(t, err)
	require.Equal(t, channel, out)
	require.Equal(t, pk, alice.ID())

	infoOut, pk, err := invite.DecryptKey(saltpack.NewKeyring(alice))
	require.NoError(t, err)
	require.Equal(t, infoOut, infoOut)
	require.Equal(t, pk, alice.ID())
}
