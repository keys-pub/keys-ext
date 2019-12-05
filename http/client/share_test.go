package client

import (
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestShare(t *testing.T) {
	clock := newClock()
	testClient := testClient(t, clock)
	cl := testClient.client
	defer testClient.closeFn()

	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)
	group, err := keys.NewKeyFromSeedPhrase(groupSeed, false)
	require.NoError(t, err)
	errA := testClient.ks.SaveKey(alice, true, clock.Now())
	require.NoError(t, errA)
	errG := testClient.ks.SaveKey(group, true, clock.Now())
	require.NoError(t, errG)

	b, err := cl.Shared(alice, group.ID())
	require.NoError(t, err)
	require.Nil(t, b)

	url, err := cl.Share(alice.PublicKey(), group, []byte("ok"))
	require.NoError(t, err)
	require.Equal(t, cl.url.String()+"/share/HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec/QcCryFxU6wcYxQ4DME9PP1kbq76nf2YtAqk2GwHQqfqR", url)
	out, err := cl.Shared(alice, group.ID())
	require.NoError(t, err)
	require.Equal(t, "ok", string(out))

	url2, err := cl.Share(alice.PublicKey(), group, []byte("ok2"))
	require.NoError(t, err)
	require.Equal(t, cl.url.String()+"/share/HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec/QcCryFxU6wcYxQ4DME9PP1kbq76nf2YtAqk2GwHQqfqR", url2)
	out2, err := cl.Shared(alice, group.ID())
	require.NoError(t, err)
	require.Equal(t, "ok2", string(out2))

	err = cl.DeleteShare(alice.PublicKey(), group)
	require.NoError(t, err)

	err = cl.DeleteShare(alice.PublicKey(), group)
	require.IsType(t, err, ErrResponse{})
	require.Equal(t, 404, err.(ErrResponse).StatusCode)
	require.EqualError(t, err, "404 share not found")
}
