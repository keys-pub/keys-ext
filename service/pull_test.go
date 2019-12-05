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

	keysResp, err := aliceService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(keysResp.Keys))
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", keysResp.Keys[0].KID)

	// Bob
	bobService, bobCloseFn := testServiceFire(t, fi, clock)
	defer bobCloseFn()
	testAuthSetup(t, bobService, bob, true, "bob")

	// Bob pull (all)
	pullResp, err := bobService.Pull(ctx, &PullRequest{All: true})
	require.NoError(t, err)
	require.Equal(t, 2, len(pullResp.KIDs))
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", pullResp.KIDs[0])
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", pullResp.KIDs[1])
	keysResp, err = bobService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(keysResp.Keys))
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", keysResp.Keys[0].KID)
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", keysResp.Keys[1].KID)

	// Alice (pull bob KID)
	pullResp, err = aliceService.Pull(ctx, &PullRequest{KID: "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x"})
	require.NoError(t, err)
	require.Equal(t, 1, len(pullResp.KIDs))
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", pullResp.KIDs[0])
	keysResp, err = aliceService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(keysResp.Keys))
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", keysResp.Keys[0].KID)
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", keysResp.Keys[1].KID)

	// Charlie
	charlieService, charlieCloseFn := testServiceFire(t, fi, clock)
	defer charlieCloseFn()
	testAuthSetup(t, charlieService, charlie, true, "charlie")

	// Charlie pull (alice)
	pullResp, err = charlieService.Pull(ctx, &PullRequest{User: "alice@test"})
	require.NoError(t, err)
	require.Equal(t, 1, len(pullResp.KIDs))
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", pullResp.KIDs[0])
}
