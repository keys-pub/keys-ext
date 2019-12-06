package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPull(t *testing.T) {
	clock := newClock()
	fi := testFire(t, clock)
	ctx := context.TODO()

	// Alice
	aliceService, aliceCloseFn := testServiceFire(t, fi, clock)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService, alice, true, "alice")

	respKeys, err := aliceService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(respKeys.Keys))
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", respKeys.Keys[0].KID)

	// Alice pull (default)
	resp, err := aliceService.Pull(ctx, &PullRequest{})
	require.NoError(t, err)
	require.Equal(t, []string{"ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg"}, resp.KIDs)

	// Bob
	bobService, bobCloseFn := testServiceFire(t, fi, clock)
	defer bobCloseFn()
	testAuthSetup(t, bobService, bob, true, "bob")

	// Bob pull (all)
	resp, err = bobService.Pull(ctx, &PullRequest{All: true})
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.KIDs))
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.KIDs[0])
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", resp.KIDs[1])
	respKeys, err = bobService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(respKeys.Keys))
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", respKeys.Keys[0].KID)
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", respKeys.Keys[1].KID)

	// Alice (pull bob KID)
	resp, err = aliceService.Pull(ctx, &PullRequest{KID: "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x"})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.KIDs))
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", resp.KIDs[0])
	respKeys, err = aliceService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(respKeys.Keys))
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", respKeys.Keys[0].KID)
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", respKeys.Keys[1].KID)

	// Charlie
	charlieService, charlieCloseFn := testServiceFire(t, fi, clock)
	defer charlieCloseFn()
	testAuthSetup(t, charlieService, charlie, true, "charlie")

	// Charlie pull (alice)
	resp, err = charlieService.Pull(ctx, &PullRequest{User: "alice@test"})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.KIDs))
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.KIDs[0])
}
