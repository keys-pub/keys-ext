package service

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestStatusPrompts(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	resp, err := service.Status(ctx, &StatusRequest{})
	require.EqualError(t, err, "keyring is locked")

	testAuthSetup(t, service, alice)

	resp, err = service.Status(ctx, &StatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, alice.ID().String(), resp.Key.ID)
	require.Equal(t, 0, len(resp.Key.Users))
	require.True(t, resp.PromptUser)

	_, err = service.ConfigSet(ctx, &ConfigSetRequest{
		Key:   "disablePromptUser",
		Value: "1",
	})
	require.NoError(t, err)

	resp, err = service.Status(ctx, &StatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, alice.ID().String(), resp.Key.ID)
	require.Equal(t, 0, len(resp.Key.Users))
	require.False(t, resp.PromptUser)
}

func TestStatusUser(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice)
	testUserSetup(t, env, service, alice, "alice")
	testPush(t, service, alice)

	resp, err := service.Status(ctx, &StatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, alice.ID().String(), resp.Key.ID)
	require.Equal(t, 1, len(resp.Key.Users))
	require.Equal(t, UserStatusOK, resp.Key.Users[0].Status)
	require.Equal(t, int64(1234567890018), resp.Key.Users[0].VerifiedAt)
	require.False(t, resp.PromptUser)

	// Set error and update
	env.req.SetError("https://gist.github.com/alice/1", errors.Errorf("test error"))
	_, err = service.users.Update(context.TODO(), alice.ID())
	require.NoError(t, err)

	resp, err = service.Status(ctx, &StatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, alice.ID().String(), resp.Key.ID)
	require.Equal(t, 1, len(resp.Key.Users))
	require.Equal(t, UserStatusConnFailure, resp.Key.Users[0].Status)
	require.Equal(t, int64(1234567890018), resp.Key.Users[0].VerifiedAt)
	require.Equal(t, "test error", resp.Key.Users[0].Err)
}
