package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestKeyImportExport(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{Type: EdX25519})
	require.NoError(t, err)
	kid, err := keys.ParseID(genResp.KID)
	require.NoError(t, err)

	// Export
	exportResp, err := service.KeyExport(ctx, &KeyExportRequest{
		KID:      kid.String(),
		Password: "test",
	})
	require.NoError(t, err)
	require.NotEmpty(t, exportResp.Export)

	// Remove
	_, err = service.KeyRemove(ctx, &KeyRemoveRequest{KID: kid.String()})
	require.NoError(t, err)

	// Import
	importResp, err := service.KeyImport(ctx, &KeyImportRequest{
		In:       exportResp.Export,
		Password: "test",
	})
	require.NoError(t, err)
	require.Equal(t, kid.String(), importResp.KID)

	keyResp, err := service.Key(ctx, &KeyRequest{KID: kid.String()})
	require.NoError(t, err)
	require.Equal(t, kid.String(), keyResp.Key.ID)

	// Import (bob, ID)
	importResp, err = service.KeyImport(ctx, &KeyImportRequest{
		In: []byte(bob.ID().String()),
	})
	require.NoError(t, err)
	require.Equal(t, bob.ID().String(), importResp.KID)

	// Import (error)
	_, err = service.KeyImport(ctx, &KeyImportRequest{In: []byte{}})
	require.EqualError(t, err, "failed to import key: failed to decrypt saltpack encoded key: failed to decrypt with a password: not enough bytes")
}
