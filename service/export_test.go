package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestKeyExport(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)

	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{Type: EdX25519})
	require.NoError(t, err)
	kid, err := keys.ParseID(genResp.KID)
	require.NoError(t, err)

	resp, err := service.KeyExport(ctx, &KeyExportRequest{
		KID:      kid.String(),
		Password: "invalid",
	})
	require.EqualError(t, err, "invalid password")

	resp, err = service.KeyExport(ctx, &KeyExportRequest{
		KID:      kid.String(),
		Password: "testpassword",
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Export)

	out, err := keys.DecodeKeyFromSaltpack(string(resp.Export), "testpassword", false)
	require.NoError(t, err)
	require.Equal(t, kid, out.ID())
}
