package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSigchain(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetup(t, env, service, alice, "alice")

	sc, err := service.scs.Sigchain(alice.ID())
	require.NoError(t, err)
	require.Equal(t, 1, len(sc.Statements()))
	st := sc.Statements()[0]
	rst := statementToRPC(st)
	out := statementFromRPC(rst)
	require.Equal(t, st.Bytes(), out.Bytes())

	ctx := context.TODO()
	resp, err := service.Sigchain(ctx, &SigchainRequest{
		KID: alice.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.Key.ID)
	require.Equal(t, 1, len(resp.Statements))

	rsc, err := sigchainFromRPC(resp.Key.ID, resp.Statements, alice.PublicKey())
	require.NoError(t, err)
	require.Equal(t, 1, len(rsc.Statements()))
}
