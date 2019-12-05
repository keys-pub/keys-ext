package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service, alice, true, "alice")

	testRecoverKey(t, service, bob, false, "bob")
	testRecoverKey(t, service, charlie, true, "")
	testRemoveKey(t, service, charlie)

	resp, err := service.Search(ctx, &SearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Results))
	// Alice
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Results[0].KID)
	require.Equal(t, 1, len(resp.Results[0].Users))
	require.Equal(t, "alice", resp.Results[0].Users[0].Name)
	require.Equal(t, PrivateKeyType, resp.Results[0].Type)
	require.True(t, resp.Results[0].Saved)
	// Charlie
	require.Equal(t, "HBtyNnL4mJYQj2QtAb982yokS1Fgy5VYj7Bh5NFBkycS", resp.Results[1].KID)
	require.Equal(t, PublicKeyType, resp.Results[1].Type)
	require.False(t, resp.Results[1].Saved)
}
