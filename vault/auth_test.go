package vault_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	vlt := newTestVault(t, false)
	testAuth(t, vlt)
}

func testAuth(t *testing.T, vlt *vault.Vault) {
	var err error

	status, err := vlt.Status()
	require.NoError(t, err)
	require.Equal(t, vault.Setup, status)

	salt := bytes.Repeat([]byte{0x01}, 32)
	key, err := keys.KeyForPassword("password123", salt)
	require.NoError(t, err)

	// Unlock (error)
	_, err = vlt.Unlock(key)
	require.EqualError(t, err, "invalid vault auth")

	_, err = keys.KeyForPassword("", salt)
	require.EqualError(t, err, "empty password")

	// Setup
	provision := vault.NewProvision(vault.UnknownAuth)
	err = vlt.Setup(key, provision)
	require.NoError(t, err)

	status, err = vlt.Status()
	require.NoError(t, err)
	require.Equal(t, vault.Unlocked, status)

	// Setup (again)
	err = vlt.Setup(key, provision)
	require.EqualError(t, err, "vault is already setup")

	// Lock
	err = vlt.Lock()
	require.NoError(t, err)

	status, err = vlt.Status()
	require.NoError(t, err)
	require.Equal(t, vault.Locked, status)

	_, err = vlt.Unlock(key)
	require.NoError(t, err)

	status, err = vlt.Status()
	require.NoError(t, err)
	require.Equal(t, vault.Unlocked, status)

	// Create item
	item := vault.NewItem("key1", []byte("secret"), "", time.Now())
	err = vlt.Set(item)
	require.NoError(t, err)

	item, err = vlt.Get("key1")
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, "key1", item.ID)
	require.Equal(t, []byte("secret"), item.Data)

	// Lock
	err = vlt.Lock()
	require.NoError(t, err)

	// Check provisions
	mds, err := vlt.Provisions()
	require.NoError(t, err)
	require.Equal(t, 1, len(mds))
	require.Equal(t, provision.ID, mds[0].ID)

	// Provision
	provision2 := vault.NewProvision(vault.UnknownAuth)
	key2, err := keys.KeyForPassword("diffpassword", salt)
	require.NoError(t, err)
	err = vlt.Provision(key2, provision2)
	require.EqualError(t, err, "vault is locked")
	_, err = vlt.Unlock(key)
	require.NoError(t, err)
	err = vlt.Provision(key2, provision2)
	require.NoError(t, err)

	// Test both succeed
	err = vlt.Lock()
	require.NoError(t, err)
	_, err = vlt.Unlock(key)
	require.NoError(t, err)
	err = vlt.Lock()
	require.NoError(t, err)
	_, err = vlt.Unlock(key2)
	require.NoError(t, err)

	// Deprovision
	ok, err := vlt.Deprovision(provision2.ID, false)
	require.NoError(t, err)
	require.True(t, ok)

	_, err = vlt.Unlock(key2)
	require.EqualError(t, err, "invalid vault auth")

	// Test wrong password
	wrongpass, err := keys.KeyForPassword("invalidpassword", salt)
	require.NoError(t, err)
	_, err = vlt.Unlock(wrongpass)
	require.EqualError(t, err, "invalid vault auth")
}

func TestAuthV0(t *testing.T) {
	vlt := vault.New(vault.NewMem(), vault.V1())

	salt := bytes.Repeat([]byte{0x01}, 32)
	key, err := keys.KeyForPassword("password123", salt)
	require.NoError(t, err)

	// Set auth the old way
	item := vault.NewItem("#auth", key[:], "", time.Now())
	b, err := item.Encrypt(key)
	require.NoError(t, err)
	err = vlt.Store().Set("#auth", b)
	require.NoError(t, err)
	err = vlt.Store().Set("#salt", salt)
	require.NoError(t, err)

	// Unlock with old auth
	_, err = vlt.Unlock(key)
	require.NoError(t, err)

	// Unlock with password
	err = vlt.UnlockWithPassword("password123", false)
	require.NoError(t, err)

	provisions, err := vlt.Provisions()
	require.NoError(t, err)
	require.Equal(t, 1, len(provisions))
	require.Equal(t, "auth.v0", provisions[0].ID)

	ok, err := vlt.Deprovision("auth.v0", true)
	require.NoError(t, err)
	require.True(t, ok)
	provisions, err = vlt.Provisions()
	require.NoError(t, err)
	require.Equal(t, 0, len(provisions))
}
