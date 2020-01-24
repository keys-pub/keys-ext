package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestKey(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	// db.SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service, alice)
	testUserSetup(t, env, service, alice, "alice")

	// Alice
	resp, err := service.Key(ctx, &KeyRequest{
		KID: alice.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.Key.ID)

	// Alice (user)
	resp, err = service.Key(ctx, &KeyRequest{
		User: "alice@github",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, alice.ID().String(), resp.Key.ID)
}

func TestKeyGenerate(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice)

	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{})
	require.NoError(t, err)

	key, err := service.parseSignKey(genResp.KID, true)
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, key.ID().String(), genResp.KID)
}

func TestKeyImportExportRemove(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice)

	// Register
	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{})
	require.NoError(t, err)
	kid, err := keys.ParseID(genResp.KID)
	require.NoError(t, err)
	key, err := service.ks.SignKey(kid)
	require.NoError(t, err)
	require.NotNil(t, key)

	// Export
	exportResp, err := service.KeyExport(ctx, &KeyExportRequest{
		KID:      key.ID().String(),
		Password: "test",
	})
	require.NoError(t, err)
	require.NotEmpty(t, exportResp.Export)

	// Remove
	_, err = service.KeyRemove(ctx, &KeyRemoveRequest{KID: key.ID().String()})
	require.NoError(t, err)

	// Remove (not found)
	randKey := keys.GenerateEd25519Key()
	_, err = service.KeyRemove(ctx, &KeyRemoveRequest{KID: randKey.ID().String()})
	require.EqualError(t, err, fmt.Sprintf("not found %s", randKey.ID()))

	// Import
	_, err = service.KeyImport(ctx, &KeyImportRequest{In: []byte{}})
	require.EqualError(t, err, "failed to import key: failed to parse saltpack: missing saltpack start")

	importResp, err := service.KeyImport(ctx, &KeyImportRequest{
		In:       exportResp.Export,
		Password: "test",
	})
	require.NoError(t, err)
	require.Equal(t, key.ID().String(), importResp.KID)

	keyResp, err := service.Key(ctx, &KeyRequest{KID: key.ID().String()})
	require.NoError(t, err)
	require.Equal(t, key.ID().String(), keyResp.Key.ID)
}
