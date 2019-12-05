package client

import (
	"fmt"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	clock := newClock()
	testClient := testClient(t, clock)
	cl := testClient.client
	defer testClient.closeFn()

	i := 0
	for {
		key := keys.GenerateKey()
		kid := key.ID()
		if strings.HasPrefix(kid.String(), "a") {
			continue
		}
		if strings.HasPrefix(kid.String(), "b") {
			continue
		}
		sc := keys.NewSigchain(key.SignKey().PublicKey)
		username := fmt.Sprintf("ausername%d", i)
		i++

		usr, err := keys.NewUser(kid, "test", username, "test://", 1)
		require.NoError(t, err)
		st, err := keys.GenerateUserStatement(sc, usr, key.SignKey(), clock.Now())
		require.NoError(t, err)
		err = sc.Add(st)
		require.NoError(t, err)
		err = cl.PutSigchainStatement(st)
		require.NoError(t, err)
		err = cl.Check(key)
		require.NoError(t, err)
		if i >= 10 {
			break
		}
	}

	bob, err := keys.NewKeyFromSeedPhrase(bobSeed, false)
	require.NoError(t, err)
	bobSc := keys.GenerateSigchain(bob, clock.Now())
	bobSt := bobSc.Statements()[0]
	perr := cl.PutSigchainStatement(bobSt)
	require.NoError(t, perr)

	// spew, err := keys.Spew(testCl.dst, "user-names", nil)
	// require.NoError(t, err)

	searchResp, err := cl.Search("", 0, 0)
	require.NoError(t, err)
	require.Equal(t, 11, len(searchResp.Results))
	require.Equal(t, 1, len(searchResp.Results[0].Users))
	require.Equal(t, "ausername0", searchResp.Results[0].Users[0].Name)

	searchResp, err = cl.Search("", 0, 1)
	require.NoError(t, err)
	require.Equal(t, 1, len(searchResp.Results))
	require.Equal(t, 1, len(searchResp.Results[0].Users))
	require.Equal(t, "ausername0", searchResp.Results[0].Users[0].Name)

	searchResp, err = cl.Search("ausername1", 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(searchResp.Results))
	require.Equal(t, 1, len(searchResp.Results[0].Users))
	require.Equal(t, "ausername1", searchResp.Results[0].Users[0].Name)

	searchResp, err = cl.Search("b", 0, 1)
	require.NoError(t, err)
	require.Equal(t, 0, len(searchResp.Results))

	searchResp, err = cl.Search("KNLPD1zD35F", 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(searchResp.Results))
	require.Equal(t, "KNLPD1zD35FpXxP8q2B7JEWVqeJTxYH5RQKtGgrgNAtU", searchResp.Results[0].KID.String())
}
