package service

import (
	"context"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/keys-pub/keys-ext/auth/fido2"
	"github.com/keys-pub/keys-ext/auth/mock"
	"github.com/stretchr/testify/require"
)

func testGoBin(t *testing.T) string {
	usr, err := user.Current()
	require.NoError(t, err)
	return filepath.Join(usr.HomeDir, "go", "bin")
}

func TestHMACSecretAuthOnDevice(t *testing.T) {
	if os.Getenv("TEST_FIDO2") != "1" {
		t.Skip()
	}
	pin := os.Getenv("FIDO2_PIN")
	var err error

	// SetLogger(NewLogger(DebugLevel))

	cfg, closeFn := testConfig(t, "KeysTest", "")
	defer closeFn()

	auth := newAuth(cfg)
	vlt := newTestVault(t)

	// Load plugin
	fido2Plugin, err := fido2.OpenPlugin(filepath.Join(testGoBin(t), "fido2.so"))
	require.NoError(t, err)
	auth.fas = fido2Plugin

	t.Logf("Setup")
	err = auth.setup(context.TODO(), vlt, pin, FIDO2HMACSecretAuth)
	require.NoError(t, err)

	t.Logf("Unlock")
	token, err := auth.unlock(context.TODO(), vlt, pin, FIDO2HMACSecretAuth, "test")
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestHMACSecretAuth(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// vault.SetLogger(NewLogger(DebugLevel))
	var err error

	cfg, closeFn := testConfig(t, "KeysTest", "")
	defer closeFn()
	auth := newAuth(cfg)
	vlt := newTestVault(t)
	pin := "12345"

	// Try without plugin
	err = auth.setup(context.TODO(), vlt, pin, FIDO2HMACSecretAuth)
	require.EqualError(t, err, "failed to setup: fido2 plugin not available")

	// Set mock plugin
	auths := mock.NewAuthServer()
	auth.fas = auths

	t.Logf("Setup")
	err = auth.setup(context.TODO(), vlt, pin, FIDO2HMACSecretAuth)
	require.NoError(t, err)

	t.Logf("Unlock")
	token, err := auth.unlock(context.TODO(), vlt, pin, FIDO2HMACSecretAuth, "test")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	mk := vlt.MasterKey()

	err = vlt.Lock()
	require.NoError(t, err)

	_, err = auth.unlock(context.TODO(), vlt, pin, FIDO2HMACSecretAuth, "test")
	require.NoError(t, err)
	require.Equal(t, mk, vlt.MasterKey())

	// Unset plugin
	auth.fas = nil

	_, err = auth.unlock(context.TODO(), vlt, pin, FIDO2HMACSecretAuth, "test")
	require.EqualError(t, err, "failed to unlock: fido2 plugin not available")

}
