package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetupGithub(t, env, service, alice, "alice")
	testPush(t, service, alice)

	testImportKey(t, service, bob)
	testUserSetupGithub(t, env, service, bob, "bob")
	testPush(t, service, bob)

	resp, err := service.UserSearch(ctx, &UserSearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Users))

	// Alice
	require.Equal(t, alice.ID().String(), resp.Users[0].KID)
	require.Equal(t, "alice", resp.Users[0].Name)
	// Charlie
	require.Equal(t, bob.ID().String(), resp.Users[1].KID)

	// TODO: More tests
}
