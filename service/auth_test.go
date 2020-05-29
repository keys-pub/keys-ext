package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys/keyring"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAuthWithPassword(t *testing.T) {
	cfg, closeFn := testConfig(t, "KeysTest", "", "mem")
	defer closeFn()
	auth, err := newAuth(cfg)
	require.NoError(t, err)
	defer auth.reset()
	ctx := context.TODO()

	// Setup needed
	status, err := auth.Keyring().Status()
	require.NoError(t, err)
	require.Equal(t, keyring.Setup, status)

	// Setup
	err = auth.setup(ctx, "password123", PasswordAuth)
	require.NoError(t, err)

	status, err = auth.Keyring().Status()
	require.NoError(t, err)
	require.Equal(t, keyring.Unlocked, status)

	token, err := auth.unlock(ctx, "password123", PasswordAuth, "test")
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, token)

	// Lock
	err = auth.lock()
	require.NoError(t, err)

	// Unlock with invalid password
	_, err = auth.unlock(ctx, "invalidpassword", PasswordAuth, "test")
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid password")
	require.Empty(t, auth.tokens)
	require.Empty(t, auth.tokens)

	// Unlock
	token, err = auth.unlock(ctx, "password123", PasswordAuth, "test")
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, token)
}

func TestAuthorize(t *testing.T) {
	cfg, closeFn := testConfig(t, "KeysTest", "", "mem")
	defer closeFn()
	auth, err := newAuth(cfg)
	require.NoError(t, err)
	defer auth.reset()

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

	// Unlock
	err = auth.setup(ctx, "password123", PasswordAuth)
	require.NoError(t, err)
	token, err := auth.unlock(ctx, "password123", PasswordAuth, "test")
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

func TestAuthLock(t *testing.T) {
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
	_, err = service.AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: password,
		Type:   PasswordAuth,
		Client: "test",
	})
	require.NoError(t, err)

	testImportKey(t, service, alice)

	_, err = service.Sign(context.TODO(), &SignRequest{Data: []byte("test"), Signer: alice.ID().String()})
	require.NoError(t, err)

	_, err = service.AuthLock(ctx, &AuthLockRequest{})
	require.NoError(t, err)
	require.Empty(t, service.auth.tokens)

	_, err = service.Sign(context.TODO(), &SignRequest{Data: []byte("test"), Signer: alice.ID().String()})
	require.EqualError(t, err, "keyring is locked")
}

func TestAuthSetup(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	_, err := service.AuthSetup(ctx, &AuthSetupRequest{Secret: "password123", Type: PasswordAuth})
	require.NoError(t, err)
}
