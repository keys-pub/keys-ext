package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSigchain(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetupGithub(t, env, service, alice, "alice")

	sc, err := service.scs.Sigchain(alice.ID())
	require.NoError(t, err)
	require.Equal(t, 1, len(sc.Statements()))
	st := sc.Statements()[0]
	rst := statementToRPC(st)
	out, err := statementFromRPC(rst)
	require.NoError(t, err)
	b, err := st.Bytes()
	require.NoError(t, err)
	b2, err := out.Bytes()
	require.NoError(t, err)
	require.Equal(t, b, b2)

	ctx := context.TODO()
	resp, err := service.Sigchain(ctx, &SigchainRequest{
		KID: alice.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.Key.ID)
	require.Equal(t, 1, len(resp.Statements))

	rsc, err := sigchainFromRPC(alice.ID(), resp.Statements)
	require.NoError(t, err)
	require.Equal(t, 1, len(rsc.Statements()))
}
