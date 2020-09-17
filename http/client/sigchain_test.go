package client_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestSigchain(t *testing.T) {
	env, closeFn := newEnv(t)
	defer closeFn()

	client := newTestClient(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	sc := keys.NewSigchain(alice.ID())
	st, err := keys.NewSigchainStatement(sc, []byte("testing1"), alice, "", env.clock.Now())
	require.NoError(t, err)
	err = sc.Add(st)
	require.NoError(t, err)
	err = client.SigchainSave(context.TODO(), st)
	require.NoError(t, err)

	st2, err := keys.NewSigchainStatement(sc, []byte("testing2"), alice, "", env.clock.Now())
	require.NoError(t, err)
	err = sc.Add(st2)
	require.NoError(t, err)
	psiErr2 := client.SigchainSave(context.TODO(), st2)
	require.NoError(t, psiErr2)

	scResp, err := client.Sigchain(context.TODO(), alice.ID())
	require.NoError(t, err)
	sc, err = scResp.Sigchain()
	require.NoError(t, err)
	require.Equal(t, 2, len(sc.Statements()))
	// require.Equal(t, tsutil.ParseMillis(1234567890011), sc.Statements()[0].CreatedAt)

	key := keys.GenerateEdX25519Key()
	scResp2, err := client.Sigchain(context.TODO(), key.ID())
	require.NoError(t, err)
	require.Nil(t, scResp2)

	st3, err := keys.NewSigchainStatement(sc, []byte("testing3"), alice, "", env.clock.Now())
	require.NoError(t, err)
	err = sc.Add(st3)
	require.NoError(t, err)
	psiErr3 := client.SigchainSave(context.TODO(), st3)
	require.NoError(t, psiErr3)

	spew := sc.Spew()
	t.Logf(spew.String())
}

func TestSigchainRetryOnConflict(t *testing.T) {
	attempt := 0
	// TODO: shorten retry delay for test
	handlerFn := func(w http.ResponseWriter, req *http.Request) bool {
		if attempt == 0 {
			attempt++
			w.WriteHeader(http.StatusConflict)
			return true
		}
		return false
	}

	env, closeFn := newEnvWithOptions(t, &envOptions{
		handlerFn: handlerFn,
	})
	defer closeFn()

	cl := newTestClient(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	saveUser(t, env, cl, alice, "alice", "github")

}
