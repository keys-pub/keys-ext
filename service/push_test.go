package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPush(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	ctx := context.TODO()

	_, err := service.Push(ctx, &PushRequest{Identity: alice.ID().String()})
	require.EqualError(t, err, "nothing to push")

	testUserSetupGithub(t, env, service, alice, "alice")

	res, err := service.users.User(ctx, "alice@github")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, alice.ID(), res.User.KID)

	resp, err := service.Push(ctx, &PushRequest{Identity: alice.ID().String()})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.KID)

	users, err := service.searchUsersRemote(ctx, "alice@github", 1)
	require.NoError(t, err)
	require.Equal(t, 1, len(users))
	require.Equal(t, alice.ID(), users[0].KID)
}
