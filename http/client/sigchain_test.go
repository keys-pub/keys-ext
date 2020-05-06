package client

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestSigchain(t *testing.T) {
	env := testEnv(t, logger)
	defer env.closeFn()

	ks := keys.NewMemStore(true)
	client := testClient(t, env, ks)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	sc := keys.NewSigchain(alice.ID())
	st, err := keys.NewSigchainStatement(sc, []byte("testing1"), alice, "", env.clock.Now())
	require.NoError(t, err)
	err = sc.Add(st)
	require.NoError(t, err)
	err = client.PutSigchainStatement(context.TODO(), st)
	require.NoError(t, err)

	st2, err := keys.NewSigchainStatement(sc, []byte("testing2"), alice, "", env.clock.Now())
	require.NoError(t, err)
	err = sc.Add(st2)
	require.NoError(t, err)
	psiErr2 := client.PutSigchainStatement(context.TODO(), st2)
	require.NoError(t, psiErr2)

	scResp, err := client.Sigchain(context.TODO(), alice.ID())
	require.NoError(t, err)
	sc, err = scResp.Sigchain()
	require.NoError(t, err)
	require.Equal(t, 2, len(sc.Statements()))
	// require.Equal(t, util.TimeFromMillis(1234567890011), sc.Statements()[0].CreatedAt)

	key := keys.GenerateEdX25519Key()
	scResp2, err := client.Sigchain(context.TODO(), key.ID())
	require.NoError(t, err)
	require.Nil(t, scResp2)

	st3, err := keys.NewSigchainStatement(sc, []byte("testing3"), alice, "", env.clock.Now())
	require.NoError(t, err)
	err = sc.Add(st3)
	require.NoError(t, err)
	psiErr3 := client.PutSigchainStatement(context.TODO(), st3)
	require.NoError(t, psiErr3)

	spew, err := sc.Spew()
	require.NoError(t, err)
	logger.Infof(spew.String())
}
