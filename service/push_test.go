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
	testAuthSetup(t, service, alice, false)

	ctx := context.TODO()
	resp, err := service.Push(ctx, &PushRequest{})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.KID)
	// require.Equal(t, []string{}, resp.URLs)

	resp, err = service.Push(ctx, &PushRequest{KID: alice.ID().String()})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.KID)
	// require.Equal(t, []string{}, resp.URLs)
}
