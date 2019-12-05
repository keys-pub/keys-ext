package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()

	resp, err := service.Status(ctx, &StatusRequest{})
	require.EqualError(t, err, "keyring is locked")

	testAuthSetup(t, service, alice, false, "")

	resp, err = service.Status(ctx, &StatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Key.KID)
	require.Equal(t, 0, len(resp.Key.Users))
	require.True(t, resp.PromptPublish)
	require.True(t, resp.PromptUser)

	_, err = service.ConfigSet(ctx, &ConfigSetRequest{
		Key:   "disablePromptUser",
		Value: "1",
	})
	require.NoError(t, err)

	resp, err = service.Status(ctx, &StatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Key.KID)
	require.Equal(t, 0, len(resp.Key.Users))
	require.False(t, resp.PromptUser)
}
