package vault_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/keyring"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

func TestConvertV0(t *testing.T) {
	salt := bytes.Repeat([]byte{0x01}, 32)
	key, err := keys.KeyForPassword("password123", salt)
	require.NoError(t, err)

	// V0 auth
	kr := keyring.NewMem()
	item := vault.NewItem("#auth", key[:], "", time.Now())
	b, err := item.Encrypt(key)
	require.NoError(t, err)
	err = kr.Set("#auth", b)
	require.NoError(t, err)
	err = kr.Set("#salt", salt)
	require.NoError(t, err)

	// Vault with converted store
	vlt := vault.New(vault.NewMem())
	err = vault.ConvertKeyring(kr, vlt)
	require.NoError(t, err)

	// Unlock with key
	provision, err := vlt.Unlock(key)
	require.NoError(t, err)
	require.Equal(t, "v0", provision.ID)

	// Unlock with password
	err = vlt.UnlockWithPassword("password123", false)
	require.NoError(t, err)

	provisions, err := vlt.Provisions()
	require.NoError(t, err)
	require.Equal(t, 1, len(provisions))
	require.Equal(t, "v0", provisions[0].ID)

	inc, err := vlt.Increment(0)
	require.NoError(t, err)
	require.Equal(t, "000000000000003", inc)

	paths, err := vault.Paths(vlt.Store(), "")
	require.NoError(t, err)
	expected := []string{
		"/auth/v0",
		"/config/salt",
		"/db/increment",
		"/provision/v0",
		"/push/000000000000001/auth/v0",
		"/push/000000000000002/provision/v0",
		"/push/000000000000003/config/salt",
	}
	require.Equal(t, expected, paths)
}

func TestConvertV1(t *testing.T) {
	salt := bytes.Repeat([]byte{0x01}, 32)
	key, err := keys.KeyForPassword("password123", salt)
	require.NoError(t, err)

	// V1 auth
	kr := keyring.NewMem()
	item := vault.NewItem("#auth-0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29", key[:], "", time.Now())
	b, err := item.Encrypt(key)
	require.NoError(t, err)
	err = kr.Set("#auth-0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29", b)
	require.NoError(t, err)
	err = kr.Set("#salt", salt)
	require.NoError(t, err)
	provision := &vault.Provision{
		ID:        "0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		Type:      vault.PasswordAuth,
		CreatedAt: time.Now(),
	}
	b, err = msgpack.Marshal(provision)
	require.NoError(t, err)
	err = kr.Set("#provision-0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29", b)
	require.NoError(t, err)
	b2, err := item.Encrypt(key)
	require.NoError(t, err)
	err = kr.Set("key1", b2)
	require.NoError(t, err)

	// Vault with converted store
	vlt := vault.New(vault.NewMem())
	err = vault.ConvertKeyring(kr, vlt)
	require.NoError(t, err)

	// Unlock with old auth
	provision, err = vlt.Unlock(key)
	require.NoError(t, err)
	require.Equal(t, "0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29", provision.ID)

	// Unlock with password
	err = vlt.UnlockWithPassword("password123", false)
	require.NoError(t, err)

	provisions, err := vlt.Provisions()
	require.NoError(t, err)
	require.Equal(t, 1, len(provisions))
	require.Equal(t, "0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29", provisions[0].ID)

	inc, err := vlt.Increment(0)
	require.NoError(t, err)
	require.Equal(t, "000000000000004", inc)

	paths, err := vault.Paths(vlt.Store(), "")
	require.NoError(t, err)
	expected := []string{
		"/auth/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/config/salt",
		"/db/increment",
		"/item/key1",
		"/provision/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/push/000000000000001/auth/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/push/000000000000002/provision/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/push/000000000000003/config/salt",
		"/push/000000000000004/item/key1",
	}
	require.Equal(t, expected, paths)
}

func TestConvertID(t *testing.T) {
	require.Equal(t, "v0", vault.ConvertID("#auth"))
	require.Equal(t, "123", vault.ConvertID("#auth-123"))
	require.Equal(t, "123", vault.ConvertID("123"))
}
