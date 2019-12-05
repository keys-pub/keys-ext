package client

import (
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	clock := newClock()
	testClient := testClient(t, clock)
	cl := testClient.client
	defer testClient.closeFn()

	key := keys.GenerateKey()
	spk := key.PublicKey().SignPublicKey()
	kid := key.ID()
	sc := keys.NewSigchain(spk)

	usr, err := keys.NewUser(kid, "test", "testuser", "test://", 1)
	require.NoError(t, err)
	st, err := keys.GenerateUserStatement(sc, usr, key.SignKey(), clock.Now())
	require.NoError(t, err)
	err = sc.Add(st)
	require.NoError(t, err)
	perr := cl.PutSigchainStatement(st)
	require.NoError(t, perr)

	err = cl.Check(key)
	require.NoError(t, err)
}
