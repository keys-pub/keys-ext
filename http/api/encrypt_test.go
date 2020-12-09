package api_test

import (
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/saltpack"
	"github.com/stretchr/testify/require"
)

func TestEncrypt(t *testing.T) {
	info := &api.ChannelInfo{Name: "test"}

	b, err := api.Encrypt(info, alice, bob.ID())
	require.NoError(t, err)

	var out api.ChannelInfo
	pk, err := api.Decrypt(b, &out, saltpack.NewKeyring(bob))
	require.NoError(t, err)
	require.Equal(t, info, &out)
	require.Equal(t, pk, alice.ID())
}
