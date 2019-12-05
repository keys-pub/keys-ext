package client

import (
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestSigchain(t *testing.T) {
	clock := newClock()
	testClient := testClient(t, clock)
	cl := testClient.client
	defer testClient.closeFn()

	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)
	aliceSpk := alice.PublicKey().SignPublicKey()
	aliceID := alice.ID()

	sc := keys.NewSigchain(aliceSpk)
	st, err := keys.GenerateStatement(sc, []byte("testing1"), alice.SignKey(), "", clock.Now())
	require.NoError(t, err)
	err = sc.Add(st)
	require.NoError(t, err)
	err = cl.PutSigchainStatement(st)
	require.NoError(t, err)

	st2, err := keys.GenerateStatement(sc, []byte("testing2"), alice.SignKey(), "", clock.Now())
	require.NoError(t, err)
	err = sc.Add(st2)
	require.NoError(t, err)
	psiErr2 := cl.PutSigchainStatement(st2)
	require.NoError(t, psiErr2)

	scResp, err := cl.Sigchain(aliceID)
	require.NoError(t, err)
	sc, err = scResp.Sigchain()
	require.NoError(t, err)
	require.Equal(t, 2, len(sc.Statements()))
	// require.Equal(t, keys.TimeFromMillis(1234567890011), sc.Statements()[0].CreatedAt)

	randID := keys.RandID()
	scResp2, err := cl.Sigchain(randID)
	require.NoError(t, err)
	require.Nil(t, scResp2)

	resp3, err := cl.Sigchains("")
	require.NoError(t, err)
	require.Equal(t, 2, len(resp3.Statements))
	require.Equal(t, st.KID, resp3.Statements[0].KID)
	require.Equal(t, st2.KID, resp3.Statements[1].KID)
	// require.Equal(t, keys.TimeFromMillis(1234567890011), resp3.MetadataFor(resp3.Statements[0]).CreatedAt)

	st3, err := keys.GenerateStatement(sc, []byte("testing3"), alice.SignKey(), "", clock.Now())
	require.NoError(t, err)
	err = sc.Add(st3)
	require.NoError(t, err)
	psiErr3 := cl.PutSigchainStatement(st3)
	require.NoError(t, psiErr3)

	resp4, err := cl.Sigchains(resp3.Version)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp4.Statements))
	require.Equal(t, st2.KID, resp4.Statements[0].KID)
	require.Equal(t, st3.KID, resp4.Statements[1].KID)

	spew, err := sc.Spew()
	require.NoError(t, err)
	logger.Infof(spew.String())
}
