package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPush(t *testing.T) {
	clock := newClock()
	fi := testFire(t, clock)
	ctx := context.TODO()

	service, closeFn := testServiceFire(t, fi, clock)
	defer closeFn()
	testAuthSetup(t, service, alice, false, "")

	resp, err := service.Push(ctx, &PushRequest{})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.KID)
	// require.Equal(t, []string{}, resp.URLs)

	resp, err = service.Push(ctx, &PushRequest{KID: alice.ID().String()})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.KID)
	// require.Equal(t, []string{}, resp.URLs)
}
