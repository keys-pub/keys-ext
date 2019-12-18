package client

import (
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestShare(t *testing.T) {
	env := testEnv(t)
	defer env.closeFn()

	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)
	group, err := keys.NewKeyFromSeedPhrase(groupSeed, false)
	require.NoError(t, err)
	errA := env.ks.SaveKey(alice, true, env.clock.Now())
	require.NoError(t, errA)
	errG := env.ks.SaveKey(group, true, env.clock.Now())
	require.NoError(t, errG)

	b, err := env.client.Shared(alice, group.ID())
	require.NoError(t, err)
	require.Nil(t, b)

	url, err := env.client.Share(alice.PublicKey(), group, []byte("ok"))
	require.NoError(t, err)
	require.Equal(t, env.client.url.String()+"/share/HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec/QcCryFxU6wcYxQ4DME9PP1kbq76nf2YtAqk2GwHQqfqR", url)
	out, err := env.client.Shared(alice, group.ID())
	require.NoError(t, err)
	require.Equal(t, "ok", string(out))

	url2, err := env.client.Share(alice.PublicKey(), group, []byte("ok2"))
	require.NoError(t, err)
	require.Equal(t, env.client.url.String()+"/share/HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec/QcCryFxU6wcYxQ4DME9PP1kbq76nf2YtAqk2GwHQqfqR", url2)
	out2, err := env.client.Shared(alice, group.ID())
	require.NoError(t, err)
	require.Equal(t, "ok2", string(out2))

	err = env.client.DeleteShare(alice.PublicKey(), group)
	require.NoError(t, err)

	err = env.client.DeleteShare(alice.PublicKey(), group)
	require.IsType(t, err, ErrResponse{})
	require.Equal(t, 404, err.(ErrResponse).StatusCode)
	require.EqualError(t, err, "404 share not found")
}
