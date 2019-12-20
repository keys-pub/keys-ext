package service

import (
	"context"
	"encoding/hex"
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

	testAuthSetup(t, service, alice, false)
	testUserSetup(t, env, service, alice.ID(), "alice", false)

	// Alice (unpublished, check, update)
	resp, err := service.Key(ctx, &KeyRequest{
		KID:    alice.ID().String(),
		Update: true,
	})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.Key.KID)
	require.Equal(t, "b4c34364cc249280c859da12aaa359a8871efb179150798f3fe7e1ae8a5d0dba", hex.EncodeToString(resp.Key.SignPublicKey))
	require.Equal(t, "6380e8711b77937f1afb0daf821d9d9be069e3d7739bd13b044e28ccb9a8363d", hex.EncodeToString(resp.Key.BoxPublicKey))
	require.Equal(t, int64(1234567890001), resp.Key.CreatedAt)
	require.Equal(t, int64(0), resp.Key.PublishedAt)
	require.Equal(t, int64(1234567890002), resp.Key.SavedAt)

	// Alice
	resp, err = service.Key(ctx, &KeyRequest{
		KID:       alice.ID().String(),
		SkipCheck: true,
	})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.Key.KID)
	require.Equal(t, int64(1234567890001), resp.Key.CreatedAt)
	require.Equal(t, int64(0), resp.Key.PublishedAt)
	require.Equal(t, int64(1234567890002), resp.Key.SavedAt)

	testPushKey(t, service, alice)

	// Alice (published)
	resp, err = service.Key(ctx, &KeyRequest{
		KID: alice.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.Key.KID)
	require.Equal(t, int64(1234567890001), resp.Key.CreatedAt)
	require.Equal(t, int64(1234567890009), resp.Key.PublishedAt)
	require.Equal(t, int64(1234567890002), resp.Key.SavedAt)

	// Alice (user)
	resp, err = service.Key(ctx, &KeyRequest{
		User: "alice@github",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, alice.ID().String(), resp.Key.KID)
	require.Equal(t, int64(1234567890001), resp.Key.CreatedAt)
	require.Equal(t, int64(1234567890009), resp.Key.PublishedAt)
	require.Equal(t, int64(1234567890002), resp.Key.SavedAt)

	testRecoverKey(t, service, bob, true)
	testUserSetup(t, env, service, bob.ID(), "bob", true)

	// Bob (local)
	resp, err = service.Key(ctx, &KeyRequest{
		KID: bob.ID().String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, bob.ID().String(), resp.Key.KID)
	require.Equal(t, int64(1234567890030), resp.Key.CreatedAt)
	require.Equal(t, int64(0), resp.Key.PublishedAt)
	require.Equal(t, int64(1234567890031), resp.Key.SavedAt)
	require.True(t, resp.Key.Saved)

	testPullKey(t, service, bob)

	// Bob (update)
	resp, err = service.Key(ctx, &KeyRequest{
		KID:       bob.ID().String(),
		SkipCheck: true,
		Update:    true,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, bob.ID().String(), resp.Key.KID)
	require.Equal(t, int64(1234567890030), resp.Key.CreatedAt)
	require.Equal(t, int64(1234567890032), resp.Key.PublishedAt)
	require.Equal(t, int64(1234567890031), resp.Key.SavedAt)
	require.True(t, resp.Key.Saved)
}

func TestKeyGenerate(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, false)

	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{})
	require.NoError(t, err)

	key, err := service.parseKey(genResp.KID)
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, key.ID().String(), genResp.KID)
}

func TestKeyBackupRecover(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, false)

	// Register
	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{})
	require.NoError(t, err)
	kid, err := keys.ParseID(genResp.KID)
	require.NoError(t, err)
	key, err := service.ks.Key(kid)
	require.NoError(t, err)
	require.NotNil(t, key)
	seedPhrase := keys.SeedPhrase(key)

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

	keyResp, err := service.Key(ctx, &KeyRequest{KID: key.ID().String()})
	require.EqualError(t, err, fmt.Sprintf("key not found %s", key.ID()))

	// Remove (not found)
	randKey := keys.GenerateKey()
	_, err = service.KeyRemove(ctx, &KeyRemoveRequest{KID: randKey.ID().String(), SeedPhrase: keys.SeedPhrase(randKey)})
	require.EqualError(t, err, fmt.Sprintf("key not found %s", randKey.ID()))

	// Recover
	_, err = service.KeyRecover(ctx, &KeyRecoverRequest{SeedPhrase: ""})
	require.EqualError(t, err, "no seed phrase specified")

	_, err = service.KeyRecover(ctx, &KeyRecoverRequest{SeedPhrase: "foo"})
	require.EqualError(t, err, "invalid recovery phrase")

	recResp, err := service.KeyRecover(ctx, &KeyRecoverRequest{SeedPhrase: seedPhrase})
	require.NoError(t, err)
	require.Equal(t, key.ID().String(), recResp.KID)

	keyResp, err = service.Key(ctx, &KeyRequest{KID: key.ID().String()})
	require.NoError(t, err)
	require.Equal(t, key.ID().String(), keyResp.Key.KID)
}

func TestKeyShare(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	ctx := context.TODO()

	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService, alice, true)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService, bob, true)

	testRecoverKey(t, aliceService, group, true)

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

	key, err := bobService.ks.Key(group.ID())
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, group.ID().String(), key.ID().String())
}
