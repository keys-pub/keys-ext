package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLookup(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// saltpack.SetLogger(NewLogger(DebugLevel))
	// client.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)
	testUserSetupGithub(t, env, aliceService, alice, "alice")

	// Bob service
	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testUserSetupGithub(t, env, bobService, bob, "bob")
	testPull(t, bobService, "alice@github")

	ctx := context.TODO()
	kid, err := bobService.lookup(ctx, "alice@github", nil)
	require.NoError(t, err)
	require.Equal(t, alice.ID(), kid)

	_, err = bobService.edx25519Key(kid)
	require.EqualError(t, err, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 not found")
}
