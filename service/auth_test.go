package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAuthWithPassword(t *testing.T) {
	cfg, closeFn := testConfig(t, "KeysTest", "", "mem")
	defer closeFn()
	st, err := newKeyringStore(cfg)
	require.NoError(t, err)
	auth, err := newAuth(cfg, st)
	require.NoError(t, err)
	defer func() { _ = auth.keyring.Reset() }()
	kr := auth.keyring
	ctx := context.TODO()

	// Setup needed
	isSetup, err := kr.IsSetup()
	require.NoError(t, err)
	require.False(t, isSetup)

	// Setup
	_, err = auth.unlock(ctx, "password123", PasswordAuth, "test", true)
	require.NoError(t, err)

	isSetup, err = kr.IsSetup()
	require.NoError(t, err)
	require.True(t, isSetup)

	authResult, err := auth.unlock(ctx, "password123", PasswordAuth, "test", false)
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, authResult.token)

	// Lock
	err = auth.lock()
	require.NoError(t, err)

	// Unlock with invalid password
	_, err = auth.unlock(ctx, "invalidpassword", PasswordAuth, "test", false)
	require.EqualError(t, err, "rpc error: code = Unauthenticated desc = invalid password")
	require.Empty(t, auth.tokens)
	require.Empty(t, auth.tokens)

	// Unlock
	authResult, err = auth.unlock(ctx, "password123", PasswordAuth, "test", false)
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, authResult.token)
}

func TestAuthorize(t *testing.T) {
	cfg, closeFn := testConfig(t, "KeysTest", "", "mem")
	defer closeFn()
	st, err := newKeyringStore(cfg)
	require.NoError(t, err)
	auth, err := newAuth(cfg, st)
	require.NoError(t, err)
	defer func() { _ = auth.keyring.Reset() }()

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
	authResult, err := auth.unlock(ctx, "password123", PasswordAuth, "test", true)
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, authResult.token)

	ctx4 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{authResult.token},
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

	password := "password123"
	_, err := service.AuthSetup(ctx, &AuthSetupRequest{
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

	setupResp, err := service.AuthSetup(ctx, &AuthSetupRequest{Secret: "password123", Type: PasswordAuth, Client: "test"})
	require.NoError(t, err)
	require.NotEmpty(t, setupResp.AuthToken)
}
