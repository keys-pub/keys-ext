package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys-ext/vault"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAuthWithPassword(t *testing.T) {
	cfg, closeFn := testConfig(t, "KeysTest", "")
	defer closeFn()
	auth := newAuth(cfg)
	vlt := newTestVault(t)

	ctx := context.TODO()

	// Setup needed
	status, err := vlt.Status()
	require.NoError(t, err)
	require.Equal(t, vault.Setup, status)

	// Setup
	err = auth.setup(ctx, vlt, "password123", PasswordAuth)
	require.NoError(t, err)

	status, err = vlt.Status()
	require.NoError(t, err)
	require.Equal(t, vault.Unlocked, status)

	// Unlock
	token, err := auth.unlock(ctx, vlt, "password123", PasswordAuth, "test")
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, token)

	// Lock
	auth.lock(vlt)

	// Unlock with invalid password
	_, err = auth.unlock(ctx, vlt, "invalidpassword", PasswordAuth, "test")
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid password")
	require.Empty(t, auth.tokens)

	// Unlock
	token, err = auth.unlock(ctx, vlt, "password123", PasswordAuth, "test")
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, token)
}

func TestAuthorize(t *testing.T) {
	var err error
	cfg, closeFn := testConfig(t, "KeysTest", "")
	defer closeFn()
	auth := newAuth(cfg)
	vlt := newTestVault(t)

	ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{})
	err = auth.authorize(ctx, "/service.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = authorization missing")

	ctx2 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{""},
	})
	err = auth.authorize(ctx2, "/service.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid token")

	ctx3 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{"badtoken"},
	})
	err = auth.authorize(ctx3, "/service.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid token")

	// Setup
	err = auth.setup(ctx, vlt, "password123", PasswordAuth)
	require.NoError(t, err)

	token, err := auth.unlock(ctx, vlt, "password123", PasswordAuth, "test")
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, token)

	ctx4 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{token},
	})
	err = auth.authorize(ctx4, "/service.Keys/SomeMethod")
	require.NoError(t, err)

	ctx5 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{"badtoken"},
	})
	err = auth.authorize(ctx5, "/service.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid token")
}

func TestGenerateToken(t *testing.T) {
	token := generateToken()
	require.NotEmpty(t, token)
}

func TestAuthUnlockLock(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	// Setup/Unlock
	var err error
	password := "password123"
	_, err = service.AuthSetup(ctx, &AuthSetupRequest{
		Secret: password,
		Type:   PasswordAuth,
	})
	require.NoError(t, err)
	_, err = service.AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: password,
		Type:   PasswordAuth,
		Client: "test",
	})
	require.NoError(t, err)

	// Unlock again
	_, err = service.AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: password,
		Type:   PasswordAuth,
		Client: "test",
	})
	require.NoError(t, err)

	testImportKey(t, service, alice)

	_, err = service.Sign(context.TODO(), &SignRequest{Data: []byte("test"), Signer: alice.ID().String()})
	require.NoError(t, err)

	items, err := service.vault.Items()
	require.NoError(t, err)
	require.Equal(t, 1, len(items))

	_, err = service.AuthLock(ctx, &AuthLockRequest{})
	require.NoError(t, err)
	require.Empty(t, service.auth.tokens)

	_, err = service.Sign(context.TODO(), &SignRequest{Data: []byte("test"), Signer: alice.ID().String()})
	require.EqualError(t, err, "vault is locked")
}

func TestPasswordChange(t *testing.T) {
	var err error
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	_, err = service.AuthSetup(ctx, &AuthSetupRequest{Secret: "password123", Type: PasswordAuth})
	require.NoError(t, err)

	_, err = service.PasswordChange(ctx, &PasswordChangeRequest{
		Old: "invalid",
		New: "newpassword",
	})
	require.EqualError(t, err, "invalid password")

	_, err = service.PasswordChange(ctx, &PasswordChangeRequest{
		Old: "",
		New: "newpassword",
	})
	require.EqualError(t, err, "empty password")

	_, err = service.PasswordChange(ctx, &PasswordChangeRequest{
		Old: "password123",
		New: "password1234",
	})
	require.NoError(t, err)

	_, err = service.AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: "password123",
		Type:   PasswordAuth,
		Client: "test",
	})
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid password")

	_, err = service.AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: "password1234",
		Type:   PasswordAuth,
		Client: "test",
	})
	require.NoError(t, err)

}

func TestUnlockMultipleClients(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	var err error
	password := "password123"
	_, err = service.AuthSetup(ctx, &AuthSetupRequest{
		Secret: password,
		Type:   PasswordAuth,
	})
	require.NoError(t, err)

	// Unlock app
	app, err := service.AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: password,
		Type:   PasswordAuth,
		Client: "app",
	})
	require.NoError(t, err)

	// Unlock CLI
	cli, err := service.AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: password,
		Type:   PasswordAuth,
		Client: "cli",
	})
	require.NoError(t, err)

	// Check tokens
	err = service.auth.checkToken(app.AuthToken)
	require.NoError(t, err)
	err = service.auth.checkToken(cli.AuthToken)
	require.NoError(t, err)

	// Lock
	_, err = service.AuthLock(ctx, &AuthLockRequest{})
	require.NoError(t, err)

	err = service.auth.checkToken(app.AuthToken)
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid token")
	err = service.auth.checkToken(cli.AuthToken)
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid token")

	require.False(t, service.db.IsOpen())
}
