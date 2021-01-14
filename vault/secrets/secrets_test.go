package secrets_test

import (
	"testing"

	"github.com/keys-pub/keys-ext/vault/secrets"
	"github.com/stretchr/testify/require"
)

func TestSecretSave(t *testing.T) {
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	secret := &secrets.Secret{
		ID:       "Ibgoe3sXvdpxFUeR1hSUriTRdxvcoWjou80WnPiFcPC",
		Type:     secrets.PasswordType,
		Password: "mypassword",
	}

	svlt := secrets.New(vlt)

	out, updated, err := svlt.Save(secret)
	require.NoError(t, err)
	require.False(t, updated)
	require.NotNil(t, out)
	require.Equal(t, out.ID, secret.ID)
	require.Equal(t, "mypassword", secret.Password)
	out, err = svlt.Get(secret.ID)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, out.ID, secret.ID)
	require.Equal(t, "mypassword", secret.Password)

	secret.Password = "mypassword2"
	_, updated, err = svlt.Save(secret)
	require.NoError(t, err)
	require.True(t, updated)
	out, err = svlt.Get(secret.ID)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, out.ID, secret.ID)
	require.Equal(t, "mypassword2", secret.Password)
}

func TestSecrets(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()
	svlt := secrets.New(vlt)

	gabriel := secrets.NewPassword("keys.pub", "gabriel", "mypassword", "http://keys.pub")
	secret, updated, err := svlt.Save(gabriel)
	require.NoError(t, err)
	require.False(t, updated)
	require.NotNil(t, secret)
	require.Equal(t, secret.ID, gabriel.ID)
	require.Equal(t, "mypassword", gabriel.Password)

	alice := secrets.NewPassword("ok", "alice", "alicepassword", "ok")
	_, _, err = svlt.Save(alice)
	require.NoError(t, err)
	bob := secrets.NewPassword("bob.com", "bob", "bobpassword", "bob.com")
	_, _, err = svlt.Save(bob)
	require.NoError(t, err)
	charlie := secrets.NewPassword("", "charlie", "charliepassword", "")
	_, _, err = svlt.Save(charlie)
	require.NoError(t, err)

	out, err := svlt.List(secrets.WithQuery("keys.pub"))
	require.NoError(t, err)
	require.Equal(t, 1, len(out))
	require.Equal(t, gabriel.ID, out[0].ID)

	out, err = svlt.List(secrets.WithQuery("alice"))
	require.NoError(t, err)
	require.Equal(t, 1, len(out))
	require.Equal(t, alice.ID, out[0].ID)

	out, err = svlt.List(secrets.WithSort("username"))
	require.NoError(t, err)
	require.Equal(t, 4, len(out))
	require.Equal(t, alice.ID, out[0].ID)
	require.Equal(t, bob.ID, out[1].ID)
	require.Equal(t, charlie.ID, out[2].ID)
	require.Equal(t, gabriel.ID, out[3].ID)
}
