package client

import (
	"bytes"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestSigchain(t *testing.T) {
	env := testEnv(t)
	defer env.closeFn()

	alice := keys.NewEd25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	sc := keys.NewSigchain(alice.PublicKey())
	st, err := keys.GenerateStatement(sc, []byte("testing1"), alice, "", env.clock.Now())
	require.NoError(t, err)
	err = sc.Add(st)
	require.NoError(t, err)
	err = env.client.PutSigchainStatement(st)
	require.NoError(t, err)

	st2, err := keys.GenerateStatement(sc, []byte("testing2"), alice, "", env.clock.Now())
	require.NoError(t, err)
	err = sc.Add(st2)
	require.NoError(t, err)
	psiErr2 := env.client.PutSigchainStatement(st2)
	require.NoError(t, psiErr2)

	scResp, err := env.client.Sigchain(alice.ID())
	require.NoError(t, err)
	sc, err = scResp.Sigchain()
	require.NoError(t, err)
	require.Equal(t, 2, len(sc.Statements()))
	// require.Equal(t, keys.TimeFromMillis(1234567890011), sc.Statements()[0].CreatedAt)

	randID := keys.RandID("kpe")
	scResp2, err := env.client.Sigchain(randID)
	require.NoError(t, err)
	require.Nil(t, scResp2)

	resp3, err := env.client.Sigchains("")
	require.NoError(t, err)
	require.Equal(t, 2, len(resp3.Statements))
	require.Equal(t, st.KID, resp3.Statements[0].KID)
	require.Equal(t, st2.KID, resp3.Statements[1].KID)
	// require.Equal(t, keys.TimeFromMillis(1234567890011), resp3.MetadataFor(resp3.Statements[0]).CreatedAt)

	st3, err := keys.GenerateStatement(sc, []byte("testing3"), alice, "", env.clock.Now())
	require.NoError(t, err)
	err = sc.Add(st3)
	require.NoError(t, err)
	psiErr3 := env.client.PutSigchainStatement(st3)
	require.NoError(t, psiErr3)

	resp4, err := env.client.Sigchains(resp3.Version)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp4.Statements))
	require.Equal(t, st2.KID, resp4.Statements[0].KID)
	require.Equal(t, st3.KID, resp4.Statements[1].KID)

	spew, err := sc.Spew()
	require.NoError(t, err)
	logger.Infof(spew.String())
}
