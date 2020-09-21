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
	product := os.Getenv("FIDO2_PRODUCT")

	var err error

	// SetLogger(NewLogger(DebugLevel))

	env, closeFn := newEnv(t, "", "")
	defer closeFn()

	auth := newAuth(env)
	vlt := newTestVault(t)
	err = vlt.Open()
	require.NoError(t, err)
	defer vlt.Close()

	// Load plugin
	fido2Plugin, err := fido2.OpenPlugin(filepath.Join(testGoBin(t), "fido2.so"))
	require.NoError(t, err)
	auth.fas = fido2Plugin

	dev, err := findDevice(context.TODO(), fido2Plugin, product)
	require.NoError(t, err)
	require.NotNil(t, dev)

	// Setup
	err = auth.setup(context.TODO(), vlt, &AuthSetupRequest{
		Device: dev.Device.Path,
		Secret: pin,
		Type:   FIDO2HMACSecretAuth,
	})
	require.NoError(t, err)

	// Unlock
	token, err := auth.unlock(context.TODO(), vlt, &AuthUnlockRequest{
		Secret: pin,
		Type:   FIDO2HMACSecretAuth,
		Client: "test",
	})
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestHMACSecretAuth(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// vault.SetLogger(NewLogger(DebugLevel))
	var err error

	env, closeFn := newEnv(t, "", "")
	defer closeFn()
	auth := newAuth(env)
	vlt := newTestVault(t)
	err = vlt.Open()
	require.NoError(t, err)
	defer vlt.Close()
	pin := "12345"

	// Try without plugin
	err = auth.setup(context.TODO(), vlt, &AuthSetupRequest{Secret: pin, Type: FIDO2HMACSecretAuth})
	require.EqualError(t, err, "failed to setup: fido2 plugin not available")

	_, err = auth.unlock(context.TODO(), vlt, &AuthUnlockRequest{Secret: pin, Type: FIDO2HMACSecretAuth, Client: "test"})
	require.EqualError(t, err, "failed to unlock: fido2 plugin not available")

	// Set mock plugin
	auths := mock.NewAuthServer()
	auth.fas = auths

	// No device
	err = auth.setup(context.TODO(), vlt, &AuthSetupRequest{Secret: pin, Type: FIDO2HMACSecretAuth})
	require.EqualError(t, err, "failed to setup: no device specified")

	// Device not found
	err = auth.setup(context.TODO(), vlt, &AuthSetupRequest{Device: "/notfound", Secret: pin, Type: FIDO2HMACSecretAuth})
	require.EqualError(t, err, "failed to setup: device not found: /notfound")

	// Setup
	err = auth.setup(context.TODO(), vlt, &AuthSetupRequest{Device: "/dev/test", Secret: pin, Type: FIDO2HMACSecretAuth})
	require.NoError(t, err)

	// Unlock
	token, err := auth.unlock(context.TODO(), vlt, &AuthUnlockRequest{Secret: pin, Type: FIDO2HMACSecretAuth, Client: "test"})
	require.NoError(t, err)
	require.NotEmpty(t, token)

	mk := vlt.MasterKey()
	require.NotEmpty(t, mk)

	vlt.Lock()

	_, err = auth.unlock(context.TODO(), vlt, &AuthUnlockRequest{Secret: pin, Type: FIDO2HMACSecretAuth, Client: "test"})
	require.NoError(t, err)
	require.Equal(t, mk, vlt.MasterKey())
}
