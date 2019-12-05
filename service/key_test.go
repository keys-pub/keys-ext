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

	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service, alice, false, "alice")

	// Alice
	resp, err := service.Key(ctx, &KeyRequest{
		KID: "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg",
	})
	require.NoError(t, err)
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Key.KID)
	require.Equal(t, int64(1234567890001), resp.Key.CreatedAt)
	require.Equal(t, int64(0), resp.Key.PublishedAt)
	require.Equal(t, int64(1234567890002), resp.Key.SavedAt)

	testPushKey(t, service, alice)

	// Alice (published)
	resp, err = service.Key(ctx, &KeyRequest{
		KID: "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg",
	})
	require.NoError(t, err)
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Key.KID)
	require.Equal(t, int64(1234567890001), resp.Key.CreatedAt)
	require.Equal(t, int64(1234567890005), resp.Key.PublishedAt)
	require.Equal(t, int64(1234567890002), resp.Key.SavedAt)

	// Alice (user)
	resp, err = service.Key(ctx, &KeyRequest{
		User: "alice@test",
	})
	require.NoError(t, err)
	require.Equal(t, "ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg", resp.Key.KID)
	require.Equal(t, int64(1234567890001), resp.Key.CreatedAt)
	require.Equal(t, int64(1234567890005), resp.Key.PublishedAt)
	require.Equal(t, int64(1234567890002), resp.Key.SavedAt)

	testRecoverKey(t, service, bob, true, "bob")
	testRemoveKey(t, service, bob)

	// Bob (remote)
	resp, err = service.Key(ctx, &KeyRequest{
		KID:   "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x",
		Check: true,
	})
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", resp.Key.KID)
	require.Equal(t, int64(1234567890020), resp.Key.CreatedAt)
	require.Equal(t, int64(1234567890022), resp.Key.PublishedAt)
	require.Equal(t, int64(0), resp.Key.SavedAt)
	require.False(t, resp.Key.Saved)

	testPullKey(t, service, bob)

	// Bob (update)
	resp, err = service.Key(ctx, &KeyRequest{
		KID:    "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x",
		Update: true,
	})
	require.NoError(t, err)
	require.Equal(t, "6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x", resp.Key.KID)
	require.Equal(t, int64(1234567890020), resp.Key.CreatedAt)
	require.Equal(t, int64(1234567890022), resp.Key.PublishedAt)
	require.Equal(t, int64(1234567890021), resp.Key.SavedAt)
	require.True(t, resp.Key.Saved)
}

func TestKeyGenerate(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()
	testUnlock(t, service)

	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{})
	require.NoError(t, err)

	key, err := service.parseKey(genResp.KID)
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, key.ID().String(), genResp.KID)
}

func TestKeyBackupRecover(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()
	testUnlock(t, service)

	// Register
	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{})
	require.NoError(t, err)
	kid, err := keys.ParseID(genResp.KID)
	require.NoError(t, err)
	alice, err := service.ks.Key(kid)
	require.NoError(t, err)
	require.NotNil(t, alice)
	seedPhrase := keys.SeedPhrase(alice)

	// Backup
	backupResp, backupErr := service.KeyBackup(ctx, &KeyBackupRequest{
		KID: alice.ID().String(),
	})
	require.NoError(t, backupErr)
	require.Equal(t, seedPhrase, backupResp.SeedPhrase)

	// Remove
	_, removeErr := service.KeyRemove(ctx, &KeyRemoveRequest{KID: alice.ID().String()})
	require.EqualError(t, removeErr, "seed-phrase is required to remove a key, use `keys backup` to get the seed phrase")

	_, removeErr = service.KeyRemove(ctx, &KeyRemoveRequest{KID: alice.ID().String(), SeedPhrase: seedPhrase})
	require.NoError(t, removeErr)

	randKey := keys.GenerateKey()
	_, removeErr = service.KeyRemove(ctx, &KeyRemoveRequest{KID: randKey.ID().String(), SeedPhrase: keys.SeedPhrase(randKey)})
	require.EqualError(t, removeErr, fmt.Sprintf("key not found %s", randKey.ID()))

	keysResp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(keysResp.Keys))

	// Recover
	_, recErr := service.KeyRecover(ctx, &KeyRecoverRequest{SeedPhrase: ""})
	require.EqualError(t, recErr, "no seed phrase specified")

	_, recErr = service.KeyRecover(ctx, &KeyRecoverRequest{SeedPhrase: "foo"})
	require.EqualError(t, recErr, "invalid recovery phrase")

	recResp, recErr := service.KeyRecover(ctx, &KeyRecoverRequest{SeedPhrase: seedPhrase})
	require.NoError(t, recErr)
	require.Equal(t, alice.ID().String(), recResp.KID)

	keysResp, err = service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(keysResp.Keys))
	require.Equal(t, alice.ID().String(), keysResp.Keys[0].KID)
}

func TestKeyShare(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	ctx := context.TODO()

	clock := newClock()
	fi := testFire(t, clock)

	aliceService, aliceCloseFn := testServiceFire(t, fi, clock)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService, alice, true, "")

	bobService, bobCloseFn := testServiceFire(t, fi, clock)
	defer bobCloseFn()
	testAuthSetup(t, bobService, bob, true, "")

	testRecoverKey(t, aliceService, group, true, "")

	_, shareErr := aliceService.KeyShare(ctx, &KeyShareRequest{
		KID:       group.ID().String(),
		Recipient: bob.ID().String(),
	})
	require.NoError(t, shareErr)

	_, retrErr := bobService.KeyRetrieve(ctx, &KeyRetrieveRequest{
		KID:       group.ID().String(),
		Recipient: bob.ID().String(),
	})
	require.NoError(t, retrErr)

	key, err := bobService.ks.Key(keys.ID(group.ID()))
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, group.ID().String(), key.ID().String())
}

func TestKeyRemove(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()
	testUnlock(t, service)

	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{})
	require.NoError(t, err)
	backupResp, backupErr := service.KeyBackup(ctx, &KeyBackupRequest{KID: genResp.KID})
	require.NoError(t, backupErr)

	_, removeErr := service.KeyRemove(ctx, &KeyRemoveRequest{KID: genResp.KID, SeedPhrase: backupResp.SeedPhrase})
	require.NoError(t, removeErr)

	keysResp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)

	require.Equal(t, 0, len(keysResp.Keys))

	_, removeErr = service.KeyRemove(ctx, &KeyRemoveRequest{KID: alice.ID().String()})
	require.EqualError(t, removeErr, "key not found ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg")
}
