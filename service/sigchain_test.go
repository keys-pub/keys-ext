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
	testAuthSetup(t, service, alice, true)
	testUserSetup(t, env, service, alice.ID(), "alice", true)

	ctx := context.TODO()

	resp, err := service.Sigchain(ctx, &SigchainRequest{})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.Key.KID)
	require.Equal(t, 2, len(resp.Statements))

	sc, err := sigchainFromRPC(resp.Key.KID, resp.Statements)
	require.NoError(t, err)
	require.Equal(t, 2, len(sc.Statements()))
}
