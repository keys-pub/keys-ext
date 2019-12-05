package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeys(t *testing.T) {
	clock := newClock()
	fi := testFire(t, clock)
	ctx := context.TODO()

	// Alice
	aliceService, aliceCloseFn := testServiceFire(t, fi, clock)
	defer aliceCloseFn()

	testAuthSetup(t, aliceService, alice, true, "alice")

	testRecoverKey(t, aliceService, charlie, true, "")
	testRecoverKey(t, aliceService, group, true, "")

	resp, err := aliceService.Keys(ctx, &KeysRequest{
		SortField: "kid",
	})
	require.NoError(t, err)
	require.Equal(t, "kid", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 3, len(resp.Keys))
	require.Equal(t, "2d8T51ZMqoKsmyKnEAKH1NBtkjCJbjpB2PrUs6SZxsBB", resp.Keys[0].KID)
	require.Equal(t, "HBtyNnL4mJYQj2QtAb982yokS1Fgy5VYj7Bh5NFBkycS", resp.Keys[1].KID)
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Keys[2].KID)
	require.Equal(t, 1, len(resp.Keys[2].Users))
	require.Equal(t, "alice", resp.Keys[2].Users[0].Name)
	require.Equal(t, PrivateKeyType, resp.Keys[2].Type)
	require.Equal(t, int64(1234567890001), resp.Keys[2].CreatedAt)
	require.Equal(t, int64(1234567890003), resp.Keys[2].PublishedAt)
	require.Equal(t, int64(1234567890002), resp.Keys[2].SavedAt)

	// Bob
	bobService, bobCloseFn := testServiceFire(t, fi, clock)
	defer bobCloseFn()

	testAuthSetup(t, bobService, bob, true, "bob")

	pullResp, err := bobService.Pull(ctx, &PullRequest{All: true})
	require.NoError(t, err)
	require.Equal(t, 4, len(pullResp.KIDs))
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", pullResp.KIDs[0])
	require.Equal(t, "HBtyNnL4mJYQj2QtAb982yokS1Fgy5VYj7Bh5NFBkycS", pullResp.KIDs[1])
	require.Equal(t, "2d8T51ZMqoKsmyKnEAKH1NBtkjCJbjpB2PrUs6SZxsBB", pullResp.KIDs[2])
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", pullResp.KIDs[3])

	resp, err = bobService.Keys(ctx, &KeysRequest{
		SortField: "kid",
	})
	require.NoError(t, err)
	require.Equal(t, "kid", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 4, len(resp.Keys))
	require.Equal(t, "2d8T51ZMqoKsmyKnEAKH1NBtkjCJbjpB2PrUs6SZxsBB", resp.Keys[0].KID)
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", resp.Keys[1].KID)
	require.Equal(t, 1, len(resp.Keys[1].Users))
	require.Equal(t, "bob", resp.Keys[1].Users[0].Name)
	require.Equal(t, PrivateKeyType, resp.Keys[1].Type)
	require.Equal(t, "HBtyNnL4mJYQj2QtAb982yokS1Fgy5VYj7Bh5NFBkycS", resp.Keys[2].KID)
	require.Equal(t, PublicKeyType, resp.Keys[2].Type)
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Keys[3].KID)
	require.Equal(t, PublicKeyType, resp.Keys[3].Type)

	resp, err = bobService.Keys(ctx, &KeysRequest{
		SortField:     "kid",
		SortDirection: SortDesc,
	})
	require.NoError(t, err)
	require.Equal(t, "kid", resp.SortField)
	require.Equal(t, SortDesc, resp.SortDirection)
	require.Equal(t, 4, len(resp.Keys))
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Keys[0].KID)
	require.Equal(t, "HBtyNnL4mJYQj2QtAb982yokS1Fgy5VYj7Bh5NFBkycS", resp.Keys[1].KID)
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", resp.Keys[2].KID)
	require.Equal(t, "2d8T51ZMqoKsmyKnEAKH1NBtkjCJbjpB2PrUs6SZxsBB", resp.Keys[3].KID)

	resp, err = bobService.Keys(ctx, &KeysRequest{
		SortField: "user",
	})
	require.NoError(t, err)
	require.Equal(t, "user", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 4, len(resp.Keys))
	// 0: alice
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Keys[0].KID)
	require.Equal(t, 1, len(resp.Keys[0].Users))
	require.Equal(t, "alice", resp.Keys[0].Users[0].Name)
	// 1: bob
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", resp.Keys[1].KID)
	require.Equal(t, 1, len(resp.Keys[1].Users))
	require.Equal(t, "bob", resp.Keys[1].Users[0].Name)
	// 2: group
	require.Equal(t, "2d8T51ZMqoKsmyKnEAKH1NBtkjCJbjpB2PrUs6SZxsBB", resp.Keys[2].KID)
	require.Equal(t, 0, len(resp.Keys[2].Users))
	// 3: charlie
	require.Equal(t, "HBtyNnL4mJYQj2QtAb982yokS1Fgy5VYj7Bh5NFBkycS", resp.Keys[3].KID)
	require.Equal(t, 0, len(resp.Keys[3].Users))

	resp, err = bobService.Keys(ctx, &KeysRequest{
		SortField:     "user",
		SortDirection: SortDesc,
	})
	require.NoError(t, err)
	require.Equal(t, "user", resp.SortField)
	require.Equal(t, SortDesc, resp.SortDirection)
	require.Equal(t, 4, len(resp.Keys))
	// 0: bob
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", resp.Keys[0].KID)
	require.Equal(t, 1, len(resp.Keys[0].Users))
	require.Equal(t, "bob", resp.Keys[0].Users[0].Name)
	// 1: alice
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Keys[1].KID)
	require.Equal(t, 1, len(resp.Keys[1].Users))
	require.Equal(t, "alice", resp.Keys[1].Users[0].Name)
	// 2: group
	require.Equal(t, "HBtyNnL4mJYQj2QtAb982yokS1Fgy5VYj7Bh5NFBkycS", resp.Keys[2].KID)
	// 3: charlie
	require.Equal(t, "2d8T51ZMqoKsmyKnEAKH1NBtkjCJbjpB2PrUs6SZxsBB", resp.Keys[3].KID)

	resp, err = bobService.Keys(ctx, &KeysRequest{
		SortField: "type",
	})
	require.NoError(t, err)
	require.Equal(t, "type", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 4, len(resp.Keys))
	// 0: bob
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", resp.Keys[0].KID)
	require.Equal(t, 1, len(resp.Keys[0].Users))
	require.Equal(t, "bob", resp.Keys[0].Users[0].Name)
	// 1: alice
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Keys[1].KID)
	require.Equal(t, 1, len(resp.Keys[1].Users))
	require.Equal(t, "alice", resp.Keys[1].Users[0].Name)
	// 2: group
	require.Equal(t, "2d8T51ZMqoKsmyKnEAKH1NBtkjCJbjpB2PrUs6SZxsBB", resp.Keys[2].KID)
	require.Equal(t, 0, len(resp.Keys[2].Users))
	// 3: charlie
	require.Equal(t, "HBtyNnL4mJYQj2QtAb982yokS1Fgy5VYj7Bh5NFBkycS", resp.Keys[3].KID)
	require.Equal(t, 0, len(resp.Keys[3].Users))
}

func TestKeysMissingSigchain(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service, alice, true, "alice")

	_, err := service.scs.DeleteSigchain(alice.ID())
	require.NoError(t, err)

	resp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Keys))
}
