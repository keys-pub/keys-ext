package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/stretchr/testify/require"
)

func newTestVault(t *testing.T, unlocked bool) *vault.Vault {
	v := vault.New(vault.NewMem())
	if unlocked {
		v.SetMasterKey(keys.Rand32())
	}
	return v
}

func testVaultPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.vdb", keys.RandFileName()))
}

func newTestVaultDB(t *testing.T) (*vault.DB, func()) {
	db := vault.NewDB()
	path := testVaultPath()
	close := func() {
		db.Close()
		_ = os.RemoveAll(path)
	}
	err := db.OpenAtPath(path)
	require.NoError(t, err)
	return db, close
}
