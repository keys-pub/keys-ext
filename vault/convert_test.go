package vault_test

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/encoding"
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
	st := vault.NewMem()
	vlt := vault.New(st)
	converted, err := vault.ConvertKeyring(kr, vlt)
	require.NoError(t, err)
	require.True(t, converted)

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

	paths, err := vaultPaths(vlt, "")
	require.NoError(t, err)
	expected := []string{
		"/auth/v0",
		"/config/salt",
		"/provision/v0",
		"/push/000000000000001/auth/v0",
		"/push/000000000000002/provision/v0",
		"/push/000000000000003/config/salt",
		"/sync/push",
		"/sync/rsalt",
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
	st := vault.NewMem()
	vlt := vault.New(st)
	converted, err := vault.ConvertKeyring(kr, vlt)
	require.NoError(t, err)
	require.True(t, converted)

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

	paths, err := vaultPaths(vlt, "")
	require.NoError(t, err)
	expected := []string{
		"/auth/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/config/salt",
		"/item/key1",
		"/provision/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/push/000000000000001/auth/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/push/000000000000002/provision/0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
		"/push/000000000000003/config/salt",
		"/push/000000000000004/item/key1",
		"/sync/push",
		"/sync/rsalt",
	}
	require.Equal(t, expected, paths)
}

func TestConvertID(t *testing.T) {
	require.Equal(t, "v0", vault.ConvertID("#auth"))
	require.Equal(t, "123", vault.ConvertID("#auth-123"))
	require.Equal(t, "123", vault.ConvertID("123"))
}

func TestConvertBackup37(t *testing.T) {
	var err error

	// Keyring from version 0.0.37
	path := filepath.Join("testdata", "keyring-backup-1593541240195.tgz")
	kr := keyring.NewMem()
	err = keyring.Restore(path, kr)
	require.NoError(t, err)

	vlt := vault.New(vault.NewMem())
	converted, err := vault.ConvertKeyring(kr, vlt)
	require.NoError(t, err)
	require.True(t, converted)

	err = vlt.UnlockWithPassword("windows123", false)
	require.NoError(t, err)

	kys, err := vlt.Keys()
	require.NoError(t, err)
	require.Equal(t, 1, len(kys))
	require.Equal(t, keys.ID("kex1kt0wmstr4craw8d5h03uhvpyuzxudr9zypw9uzgq9nks007vy3jsasxz73"), kys[0].ID())
	require.Equal(t, "0479c4a84a474d16249dd9fba24c0ab5303a38fd62fea98c9489f7eaf71c42ebb2deedc163ae07d71db4bbe3cbb024e08dc68ca2205c5e09002ced07bfcc2465", encoding.EncodeHex(kys[0].Bytes()))

	secrets, err := vlt.Secrets()
	require.NoError(t, err)
	require.Equal(t, 1, len(secrets))
	require.Equal(t, "YIb8tyocMVunf9RTuZH36dD08rrMLGcZWGcgS68vZVu", secrets[0].ID)
	require.Equal(t, "testing", secrets[0].Name)
}

func TestConvertBackup48(t *testing.T) {
	var err error

	// Keyring from version 0.0.48
	path := filepath.Join("testdata", "keyring-backup-1593545106757.tgz")
	kr := keyring.NewMem()
	err = keyring.Restore(path, kr)
	require.NoError(t, err)

	vlt := vault.New(vault.NewMem())
	converted, err := vault.ConvertKeyring(kr, vlt)
	require.NoError(t, err)
	require.True(t, converted)

	err = vlt.UnlockWithPassword("darwin123", false)
	require.NoError(t, err)

	kys, err := vlt.Keys()
	require.NoError(t, err)
	require.Equal(t, 1, len(kys))
	require.Equal(t, keys.ID("kex1hp7507fxazu3tezfnf6mad7aw0zy5cshnyps2l8u5njqlnhev2ms58c9az"), kys[0].ID())
	require.Equal(t, "26767983d4f52553906df34729c22510ecf4d6d863f580552bb75cfdafefa7a7b87d47f926e8b915e4499a75beb7dd73c44a62179903057cfca4e40fcef962b7", encoding.EncodeHex(kys[0].Bytes()))

	secrets, err := vlt.Secrets()
	require.NoError(t, err)
	require.Equal(t, 1, len(secrets))
	require.Equal(t, "mcwuNXdxS86VRMcHjd5YrP4aR0IZQqKJfi6GTEt4c57", secrets[0].ID)
	require.Equal(t, "testing", secrets[0].Name)
}
