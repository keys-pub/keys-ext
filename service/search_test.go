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
	testUserSetup(t, env, service, alice, "alice")
	testPush(t, service, alice)

	testImportKey(t, service, bob)
	testUserSetup(t, env, service, bob, "bob")
	testPush(t, service, bob)

	resp, err := service.UserSearch(ctx, &UserSearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Results))

	// Alice
	require.Equal(t, alice.ID().String(), resp.Results[0].KID)
	require.Equal(t, "alice", resp.Results[0].User.Name)
	// Charlie
	require.Equal(t, bob.ID().String(), resp.Results[1].KID)

	// TODO: More tests
}
