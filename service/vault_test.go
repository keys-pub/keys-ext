package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/stretchr/testify/require"
)

// func newTestVaultKey(t *testing.T, clock *tsutil.Clock) (*[32]byte, *vault.Provision) {
// 	key := keys.Bytes32(bytes.Repeat([]byte{0xFF}, 32))
// 	id := encoding.MustEncode(bytes.Repeat([]byte{0xFE}, 32), encoding.Base62)
// 	provision := &vault.Provision{
// 		ID:        id,
// 		Type:      vault.UnknownAuth,
// 		CreatedAt: clock.Now(),
// 	}
// 	return key, provision
// }

func newTestVault(t *testing.T) *vault.Vault {
	return vault.New(vault.NewMem())
}

// func newTestVaultUnlocked(t *testing.T, clock *tsutil.Clock) *vault.Vault {
// 	vlt := newTestVault(t)
// 	key, provision := newTestVaultKey(t, clock)
// 	err := vlt.Setup(key, provision)
// 	require.NoError(t, err)
// 	return vlt
// }

// func testVaultPath() string {
// 	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.vdb", keys.RandFileName()))
// }

// func newTestVaultDB(t *testing.T) (*vault.DB, func()) {
// 	db := vault.NewDB()
// 	path := testVaultPath()
// 	close := func() {
// 		db.Close()
// 		_ = os.RemoveAll(path)
// 	}
// 	err := db.OpenAtPath(path)
// 	require.NoError(t, err)
// 	return db, close
// }

func TestVault(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	var err error

	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	_, err = service.VaultSync(ctx, &VaultSyncRequest{})
	require.EqualError(t, err, "failed to sync: failed to push vault: no remote set")

	testAuthSetup(t, service)

	_, err = service.VaultSync(ctx, &VaultSyncRequest{})
	require.NoError(t, err)

	status, err := service.VaultStatus(ctx, &VaultStatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, status)
	require.Equal(t, service.vault.Remote().Key.ID(), keys.ID(status.KID))

	testImportKey(t, service, alice)

	_, err = service.VaultSync(ctx, &VaultSyncRequest{})
	require.NoError(t, err)

	status, err = service.VaultStatus(ctx, &VaultStatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, status)
	rkid := status.KID

	_, err = service.VaultSync(ctx, &VaultSyncRequest{})
	require.NoError(t, err)

	status, err = service.VaultStatus(ctx, &VaultStatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, status)

	_, err = service.VaultUnsync(ctx, &VaultUnsyncRequest{})
	require.NoError(t, err)

	status, err = service.VaultStatus(ctx, &VaultStatusRequest{})
	require.NoError(t, err)
	require.Empty(t, status.KID)
	require.Empty(t, status.SyncedAt)

	// Re-sync to new vault
	_, err = service.VaultSync(ctx, &VaultSyncRequest{})
	require.NoError(t, err)

	status, err = service.VaultStatus(ctx, &VaultStatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, status)
	require.NotEqual(t, rkid, status.KID)
}
