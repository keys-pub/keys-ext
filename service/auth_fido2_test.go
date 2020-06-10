package service

import (
	"context"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/keys-pub/keys-ext/auth/fido2"
	"github.com/keys-pub/keys-ext/auth/fido2/mock"
	"github.com/stretchr/testify/require"
)

func goBin(t *testing.T) string {
	usr, err := user.Current()
	require.NoError(t, err)
	return filepath.Join(usr.HomeDir, "go", "bin")
}

func TestHMACSecretAuthOnDevice(t *testing.T) {
	if os.Getenv("FIDO2_TESTS") != "1" {
		t.Skip()
	}
	pin := os.Getenv("FIDO2_PIN")

	// SetLogger(NewLogger(DebugLevel))

	cfg, closeFn := testConfig(t, "KeysTest", "")
	defer closeFn()

	auth := newAuth(cfg)
	kr, err := newKeyring(cfg, "mem")
	require.NoError(t, err)

	// Try without plugin
	err = auth.setup(context.TODO(), kr, pin, FIDO2HMACSecretAuth)
	require.EqualError(t, err, "???")
	_, err = auth.unlock(context.TODO(), kr, pin, FIDO2HMACSecretAuth, "test")
	require.EqualError(t, err, "???")

	// Load plugin
	fido2Plugin, err := fido2.OpenPlugin(filepath.Join(goBin(t), "fido2.so"))
	require.NoError(t, err)
	auth.fas = fido2Plugin

	t.Logf("Setup")
	err = auth.setup(context.TODO(), kr, pin, FIDO2HMACSecretAuth)
	require.NoError(t, err)

	t.Logf("Unlock")
	token, err := auth.unlock(context.TODO(), kr, pin, FIDO2HMACSecretAuth, "test")
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestHMACSecretAuth(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keyring.SetLogger(NewLogger(DebugLevel))
	var err error

	cfg, closeFn := testConfig(t, "KeysTest", "")
	defer closeFn()
	auth := newAuth(cfg)
	kr, err := newKeyring(cfg, "mem")
	require.NoError(t, err)
	pin := "12345"

	// Try without plugin
	err = auth.setup(context.TODO(), kr, pin, FIDO2HMACSecretAuth)
	require.EqualError(t, err, "failed to setup: fido2 plugin not available")

	// Set mock plugin
	auths := mock.NewAuthServer()
	auth.fas = auths

	t.Logf("Setup")
	err = auth.setup(context.TODO(), kr, pin, FIDO2HMACSecretAuth)
	require.NoError(t, err)

	t.Logf("Unlock")
	token, err := auth.unlock(context.TODO(), kr, pin, FIDO2HMACSecretAuth, "test")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	mk := kr.MasterKey()

	err = kr.Lock()
	require.NoError(t, err)

	_, err = auth.unlock(context.TODO(), kr, pin, FIDO2HMACSecretAuth, "test")
	require.NoError(t, err)
	require.Equal(t, mk, kr.MasterKey())

	// Unset plugin
	auth.fas = nil

	_, err = auth.unlock(context.TODO(), kr, pin, FIDO2HMACSecretAuth, "test")
	require.EqualError(t, err, "failed to unlock: fido2 plugin not available")

}
