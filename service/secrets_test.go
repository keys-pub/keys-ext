package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecrets(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	ctx := context.TODO()

	// Alice
	service, closeFn := newTestService(t, env)
	defer closeFn()

	testAuthSetup(t, service)

	saveResp, err := service.SecretSave(ctx, &SecretSaveRequest{
		Secret: &Secret{
			Name: "testing",
			Type: PasswordSecret,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, saveResp.Secret)
	require.Equal(t, "testing", saveResp.Secret.Name)

	resp, err := service.Secrets(ctx, &SecretsRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Secrets))
}
