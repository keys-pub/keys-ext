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
	testAuthSetup(t, service, alice)
	ctx := context.TODO()

	_, err := service.Push(ctx, &PushRequest{KID: alice.ID().String()})
	require.EqualError(t, err, "nothing to push")

	testUserSetup(t, env, service, alice, "alice")

	resp, err := service.Push(ctx, &PushRequest{KID: alice.ID().String()})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.KID)
}
