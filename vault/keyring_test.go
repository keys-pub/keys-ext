package vault_test

import (
	"bytes"
	"runtime"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/keyring"
	"github.com/stretchr/testify/require"
)

func skipKeyring(t *testing.T) bool {
	if runtime.GOOS == "linux" {
		if err := keyring.CheckSystem(); err != nil {
			t.Skip()
			return true
		}
	}
	return false
}

func TestKeyringUpdate(t *testing.T) {
	if skipKeyring(t) {
		return
	}
	var err error

	kr, err := keyring.NewSystem("KeysTest")
	require.NoError(t, err)
	defer func() { _ = kr.Reset() }()

	vlt := vault.New(kr)

	testUpdate(t, vlt)
}

func TestKeyringAuth(t *testing.T) {
	if skipKeyring(t) {
		return
	}
	var err error

	kr, err := keyring.NewSystem("KeysTest")
	require.NoError(t, err)
	defer func() { _ = kr.Reset() }()

	vlt := vault.New(kr, vault.V1())
	testAuth(t, vlt)
}

// func TestKeyringSetupUnlockProvision(t *testing.T) {
// 	kr, err := keyring.NewSystem("KeysTest")
// 	require.NoError(t, err)
// 	defer func() { _ = kr.Reset() }()
// 	vlt := vault.New(kr)
// 	testSetupUnlockProvision(t, vlt)
// }

// func TestKeyringSync(t *testing.T) {
// 	kr1, err := keyring.NewSystem("KeysTest1")
// 	require.NoError(t, err)
// 	defer func() { _ = kr1.Reset() }()

// 	kr2, err := keyring.NewSystem("KeysTest2")
// 	require.NoError(t, err)
// 	defer func() { _ = kr2.Reset() }()

// 	testSync(t, kr1, kr2)
// }

func TestKeyringProtocolV1(t *testing.T) {
	var err error

	kr, err := keyring.NewSystem("KeysTest")
	require.NoError(t, err)
	defer func() { _ = kr.Reset() }()

	vlt := vault.New(kr, vault.V1())

	// Setup
	salt := bytes.Repeat([]byte{0x01}, 32)
	key, err := keys.KeyForPassword("password123", salt)
	require.NoError(t, err)
	provision := vault.NewProvision(vault.UnknownAuth)
	err = vlt.Setup(key, provision)
	require.NoError(t, err)

	// Create item
	item := vault.NewItem("testid1", []byte("testpassword"), "", time.Now())
	err = vlt.Set(item)
	require.NoError(t, err)

	_, err = vlt.Salt()
	require.NoError(t, err)

	paths, err := vault.Paths(kr, "")
	require.NoError(t, err)
	require.Equal(t, []string{
		"#auth-" + provision.ID,
		"#increment",
		"#pending-testid1-000000000000001",
		"#provision-" + provision.ID,
		"#salt",
		"testid1",
	}, paths)

	paths, err = vault.Paths(kr, "#auth")
	require.NoError(t, err)
	require.Equal(t, []string{"#auth-" + provision.ID}, paths)

	items, err := vlt.Items()
	require.NoError(t, err)
	require.Equal(t, 1, len(items))
	require.Equal(t, "testid1", items[0].ID)
}
