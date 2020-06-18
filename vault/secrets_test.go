package vault_test

import (
	"testing"

	"github.com/keys-pub/keys/secret"
	"github.com/stretchr/testify/require"
)

func TestSecretSave(t *testing.T) {
	vlt := newTestVault(t, true)

	secret := &secret.Secret{
		ID:       "Ibgoe3sXvdpxFUeR1hSUriTRdxvcoWjou80WnPiFcPC",
		Type:     secret.PasswordType,
		Password: "mypassword",
	}

	out, updated, err := vlt.SaveSecret(secret)
	require.NoError(t, err)
	require.False(t, updated)
	require.NotNil(t, out)
	require.Equal(t, out.ID, secret.ID)
	require.Equal(t, "mypassword", secret.Password)
	out, err = vlt.Secret(secret.ID)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, out.ID, secret.ID)
	require.Equal(t, "mypassword", secret.Password)

	secret.Password = "mypassword2"
	_, updated, err = vlt.SaveSecret(secret)
	require.NoError(t, err)
	require.True(t, updated)
	out, err = vlt.Secret(secret.ID)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, out.ID, secret.ID)
	require.Equal(t, "mypassword2", secret.Password)
}
