package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRand(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	_, err := service.Rand(ctx, &RandRequest{
		Encoding: BIP39,
		NumBytes: 8,
	})
	require.EqualError(t, err, "bip39 only accepts 16, 20, 24, 28, 32 bytes")

	resp, err := service.Rand(ctx, &RandRequest{
		Encoding: BIP39,
		NumBytes: 16,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Data)
}
