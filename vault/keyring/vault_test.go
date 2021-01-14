package keyring_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func NewTestVaultKey(t *testing.T, clock tsutil.Clock) (*[32]byte, *vault.Provision) {
	key := keys.Bytes32(bytes.Repeat([]byte{0xFF}, 32))
	id := encoding.MustEncode(bytes.Repeat([]byte{0xFE}, 32), encoding.Base62)
	provision := &vault.Provision{
		ID:        id,
		Type:      vault.UnknownAuth,
		CreatedAt: clock.Now(),
	}
	return key, provision
}

type StoreType string

const (
	DB  StoreType = "db"
	Mem StoreType = "mem"
)

type TestVaultOptions struct {
	Unlock bool
	Type   StoreType
	Clock  tsutil.Clock
}

func NewTestVault(t *testing.T, opts *TestVaultOptions) (*vault.Vault, func()) {
	if opts == nil {
		opts = &TestVaultOptions{}
	}
	if opts.Type == "" {
		opts.Type = Mem
	}
	if opts.Clock == nil {
		opts.Clock = tsutil.NewTestClock()
	}

	var st vault.Store
	var closeFn func()
	switch opts.Type {
	case Mem:
		st, closeFn = newTestMem(t)
	case DB:
		st, closeFn = newTestDB(t)
	}

	vlt := vault.New(st, vault.WithClock(opts.Clock))

	if opts.Unlock {
		key, provision := NewTestVaultKey(t, opts.Clock)
		err := vlt.Setup(key, provision)
		require.NoError(t, err)
		_, err = vlt.Unlock(key)
		require.NoError(t, err)
	}
	return vlt, closeFn
}

func newTestMem(t *testing.T) (vault.Store, func()) {
	mem := vault.NewMem()
	err := mem.Open()
	require.NoError(t, err)
	closeFn := func() {
		mem.Close()
	}
	return mem, closeFn
}

func newTestDB(t *testing.T) (vault.Store, func()) {
	path := testPath()
	db := vault.NewDB(path)
	err := db.Open()
	require.NoError(t, err)
	close := func() {
		err := db.Close()
		require.NoError(t, err)
		_ = os.RemoveAll(path)
	}
	return db, close
}

func testPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.vdb", keys.RandFileName()))
}
