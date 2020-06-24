package vault_test

import (
	"testing"

	"github.com/keys-pub/keys-ext/vault"
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

func TestSecrets(t *testing.T) {
	var err error
	vlt := newTestVault(t, true)

	gabriel := secret.NewPassword("gabriel", "mypassword", "keys.pub")
	out, updated, err := vlt.SaveSecret(gabriel)
	require.NoError(t, err)
	require.False(t, updated)
	require.NotNil(t, out)
	require.Equal(t, out.ID, gabriel.ID)
	require.Equal(t, "mypassword", gabriel.Password)

	alice := secret.NewPassword("alice", "alicepassword", "ok")
	_, _, err = vlt.SaveSecret(alice)
	require.NoError(t, err)
	bob := secret.NewPassword("bob", "bobpassword", "bob.com")
	_, _, err = vlt.SaveSecret(bob)
	require.NoError(t, err)
	charlie := secret.NewPassword("charlie", "charliepassword", "")
	_, _, err = vlt.SaveSecret(charlie)
	require.NoError(t, err)

	secrets, err := vlt.Secrets(vault.Secrets.Query("keys.pub"))
	require.NoError(t, err)
	require.Equal(t, 1, len(secrets))
	require.Equal(t, gabriel.ID, secrets[0].ID)

	secrets, err = vlt.Secrets(vault.Secrets.Query("alice"))
	require.NoError(t, err)
	require.Equal(t, 1, len(secrets))
	require.Equal(t, alice.ID, secrets[0].ID)

	secrets, err = vlt.Secrets(vault.Secrets.Sort("username"))
	require.NoError(t, err)
	require.Equal(t, 4, len(secrets))
	require.Equal(t, alice.ID, secrets[0].ID)
	require.Equal(t, bob.ID, secrets[1].ID)
	require.Equal(t, charlie.ID, secrets[2].ID)
	require.Equal(t, gabriel.ID, secrets[3].ID)
}
