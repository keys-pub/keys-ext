package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAuthWithPassword(t *testing.T) {
	var err error
	env, closeFn := newEnv(t, "", "")
	defer closeFn()
	auth := newAuth(env)
	vlt := newTestVault(t)
	err = vlt.Open()
	require.NoError(t, err)
	defer vlt.Close()

	ctx := context.TODO()

	// Setup needed
	status, err := vlt.Status()
	require.NoError(t, err)
	require.Equal(t, vault.SetupNeeded, status)

	// Setup
	err = auth.setup(ctx, vlt, &AuthSetupRequest{Secret: "password123", Type: PasswordAuth})
	require.NoError(t, err)

	status, err = vlt.Status()
	require.NoError(t, err)
	require.Equal(t, vault.Locked, status)

	// Unlock
	token, err := auth.unlock(ctx, vlt, &AuthUnlockRequest{Secret: "password123", Type: PasswordAuth, Client: "test"})
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, token)

	// Lock
	auth.lock(vlt)

	// Unlock with invalid password
	_, err = auth.unlock(ctx, vlt, &AuthUnlockRequest{Secret: "invalidpassword", Type: PasswordAuth, Client: "test"})
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid password")
	require.Empty(t, auth.tokens)

	// Unlock
	token, err = auth.unlock(ctx, vlt, &AuthUnlockRequest{Secret: "password123", Type: PasswordAuth, Client: "test"})
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, token)
}

func TestAuthorize(t *testing.T) {
	var err error
	env, closeFn := newEnv(t, "", "")
	defer closeFn()
	auth := newAuth(env)
	vlt := newTestVault(t)
	err = vlt.Open()
	require.NoError(t, err)
	defer vlt.Close()

	ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{})
	err = auth.authorize(ctx, "/keys.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = authorization missing")

	ctx2 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{""},
	})
	err = auth.authorize(ctx2, "/keys.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid token")

	ctx3 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{"badtoken"},
	})
	err = auth.authorize(ctx3, "/keys.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid token")

	// Setup
	err = auth.setup(ctx, vlt, &AuthSetupRequest{Secret: "password123", Type: PasswordAuth})
	require.NoError(t, err)

	token, err := auth.unlock(ctx, vlt, &AuthUnlockRequest{Secret: "password123", Type: PasswordAuth, Client: "test"})
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, token)

	ctx4 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{token},
	})
	err = auth.authorize(ctx4, "/keys.Keys/SomeMethod")
	require.NoError(t, err)

	ctx5 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{"badtoken"},
	})
	err = auth.authorize(ctx5, "/keys.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid token")
}

func TestGenerateToken(t *testing.T) {
	token := generateToken()
	require.NotEmpty(t, token)
}

func TestAuthUnlockLock(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
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
	// vault.SetLogger(NewLogger(DebugLevel))
	var err error
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	_, err = service.AuthSetup(ctx, &AuthSetupRequest{Secret: "password123", Type: PasswordAuth})
	require.NoError(t, err)

	_, err = service.AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: "password123",
		Type:   PasswordAuth,
		Client: "test",
	})
	require.NoError(t, err)

	// _, err = service.VaultSync(context.TODO(), &VaultSyncRequest{})
	// require.NoError(t, err)

	_, err = service.AuthPasswordChange(ctx, &AuthPasswordChangeRequest{
		Old: "invalid",
		New: "newpassword",
	})
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid password")

	_, err = service.AuthPasswordChange(ctx, &AuthPasswordChangeRequest{
		Old: "",
		New: "newpassword",
	})
	require.EqualError(t, err, "empty password")

	_, err = service.AuthPasswordChange(ctx, &AuthPasswordChangeRequest{
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
	service, closeFn := newTestService(t, env)
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

func TestAuthReset(t *testing.T) {
	var err error
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	_, err = service.AuthSetup(ctx, &AuthSetupRequest{Secret: "password123", Type: PasswordAuth})
	require.NoError(t, err)
	_, err = service.AuthUnlock(ctx, &AuthUnlockRequest{Secret: "password123", Type: PasswordAuth})
	require.NoError(t, err)

	_, err = service.KeyGenerate(ctx, &KeyGenerateRequest{Type: string(keys.EdX25519)})
	require.NoError(t, err)

	keysResp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(keysResp.Keys))

	_, err = service.AuthReset(ctx, &AuthResetRequest{AppName: service.env.AppName()})
	require.EqualError(t, err, "failed to reset: auth is unlocked")

	_, err = service.AuthLock(ctx, &AuthLockRequest{})
	require.NoError(t, err)

	_, err = service.AuthReset(ctx, &AuthResetRequest{AppName: "InvalidAppName"})
	require.EqualError(t, err, "failed to reset: invalid app name")

	_, err = service.AuthReset(ctx, &AuthResetRequest{AppName: service.env.AppName()})
	require.NoError(t, err)

	_, err = service.AuthSetup(ctx, &AuthSetupRequest{Secret: "password12345", Type: PasswordAuth})
	require.NoError(t, err)
	_, err = service.AuthUnlock(ctx, &AuthUnlockRequest{Secret: "password12345", Type: PasswordAuth})
	require.NoError(t, err)

	_, err = service.KeyGenerate(ctx, &KeyGenerateRequest{Type: string(keys.EdX25519)})
	require.NoError(t, err)

	keysResp, err = service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(keysResp.Keys))
}

func TestAuthSetupLocked(t *testing.T) {
	var err error
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	_, err = service.AuthSetup(ctx, &AuthSetupRequest{Secret: "password123", Type: PasswordAuth})
	require.NoError(t, err)

	_, err = service.KeyGenerate(ctx, &KeyGenerateRequest{Type: string(keys.EdX25519)})
	require.EqualError(t, err, "vault is locked")
}
