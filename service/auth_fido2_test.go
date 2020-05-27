package service

import (
	"context"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/keys-pub/keysd/auth/fido2"
	"github.com/stretchr/testify/require"
)

func goBin(t *testing.T) string {
	usr, err := user.Current()
	require.NoError(t, err)
	return filepath.Join(usr.HomeDir, "go", "bin")
}

func TestHMACSecretAuth(t *testing.T) {
	if os.Getenv("FIDO2_TESTS") != "1" {
		t.Skip()
	}
	pin := os.Getenv("FIDO2_PIN")

	SetLogger(NewLogger(DebugLevel))

	cfg, closeFn := testConfig(t, "KeysTest", "", "mem")
	defer closeFn()
	st, err := newKeyringStore(cfg)
	require.NoError(t, err)
	auth, err := newAuth(cfg, st)
	require.NoError(t, err)

	fido2Plugin, err := fido2.OpenPlugin(filepath.Join(goBin(t), "fido2.so"))
	require.NoError(t, err)
	auth.fido2 = fido2Plugin

	t.Logf("Setup")
	err = auth.setup(context.TODO(), pin, FIDO2HMACSecretAuth)
	require.NoError(t, err)

	t.Logf("Unlock")
	token, err := auth.unlock(context.TODO(), pin, FIDO2HMACSecretAuth, "test")
	require.NoError(t, err)
	require.NotEmpty(t, token)
}
