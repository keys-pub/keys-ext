package service

import (
	"testing"

	"github.com/keys-pub/keys-ext/vault"
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
