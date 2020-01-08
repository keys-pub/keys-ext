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

	key, err := service.parseKey(genResp.KID)
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, key.ID().String(), genResp.KID)
}

func TestKeyBackupRemoveRecover(t *testing.T) {
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
	seedPhrase, err := keys.BytesToPhrase(key.Seed())
	require.NoError(t, err)

	// Backup
	backupResp, err := service.KeyBackup(ctx, &KeyBackupRequest{
		KID: key.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, seedPhrase, backupResp.SeedPhrase)

	// Remove
	_, err = service.KeyRemove(ctx, &KeyRemoveRequest{KID: key.ID().String()})
	require.EqualError(t, err, "seed-phrase is required to remove a key, use `keys backup` to get the seed phrase")

	_, err = service.KeyRemove(ctx, &KeyRemoveRequest{KID: key.ID().String(), SeedPhrase: seedPhrase})
	require.NoError(t, err)

	// Remove (not found)
	randKey := keys.GenerateSignKey()
	randPhrase, err := keys.BytesToPhrase(randKey.Seed())
	require.NoError(t, err)
	_, err = service.KeyRemove(ctx, &KeyRemoveRequest{KID: randKey.ID().String(), SeedPhrase: randPhrase})
	require.EqualError(t, err, fmt.Sprintf("not found %s", randKey.ID()))

	// Recover
	_, err = service.KeyRecover(ctx, &KeyRecoverRequest{SeedPhrase: ""})
	require.EqualError(t, err, "no seed phrase specified")

	_, err = service.KeyRecover(ctx, &KeyRecoverRequest{SeedPhrase: "foo"})
	require.EqualError(t, err, "invalid recovery phrase")

	recResp, err := service.KeyRecover(ctx, &KeyRecoverRequest{SeedPhrase: seedPhrase})
	require.NoError(t, err)
	require.Equal(t, key.ID().String(), recResp.KID)

	keyResp, err := service.Key(ctx, &KeyRequest{KID: key.ID().String()})
	require.NoError(t, err)
	require.Equal(t, key.ID().String(), keyResp.Key.ID)
}
