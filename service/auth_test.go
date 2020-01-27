package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAuth(t *testing.T) {
	cfg, closeFn := testConfig(t, "")
	defer closeFn()
	auth, err := newAuth(cfg)
	require.NoError(t, err)
	defer func() { _ = auth.keyring.Reset() }()
	kr := auth.keyring

	// Setup needed
	authed, err := kr.Authed()
	require.NoError(t, err)
	require.False(t, authed)

	// Unlock (setup)
	_, err = auth.unlock("password123", "test")
	require.NoError(t, err)

	authed2, err := kr.Authed()
	require.NoError(t, err)
	require.True(t, authed2)

	token, err := auth.unlock("password123", "test")
	require.NoError(t, err)
	require.NotEmpty(t, auth.tokens)
	require.NotEmpty(t, token)

	// Lock
	auth.lock()

	// Unlock with invalid password
	_, err = auth.unlock("invalidpassword", "test")
	require.EqualError(t, err, "rpc error: code = PermissionDenied desc = invalid password")
	require.Empty(t, auth.tokens)
	require.Empty(t, auth.tokens)
}

func TestAuthorize(t *testing.T) {
	cfg, closeFn := testConfig(t, "")
	defer closeFn()
	auth, err := newAuth(cfg)
	require.NoError(t, err)
	defer func() { _ = auth.keyring.Reset() }()

	ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{})
	err = auth.authorize(ctx, "/service.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = PermissionDenied desc = authorization missing")

	ctx2 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{""},
	})
	err = auth.authorize(ctx2, "/service.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = PermissionDenied desc = invalid token")

	ctx3 := metadata.NewIncomingContext(context.TODO(), metadata.MD{
		"authorization": []string{"badtoken"},
	})
	err = auth.authorize(ctx3, "/service.Keys/SomeMethod")
	require.EqualError(t, err, "rpc error: code = PermissionDenied desc = invalid token")

	// Unlock
	token, err := auth.unlock("password123", "test")
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
	require.EqualError(t, err, "rpc error: code = PermissionDenied desc = invalid token")
}

func TestGenerateToken(t *testing.T) {
	token := generateToken()
	require.NotEmpty(t, token)
}

func TestAuthLock(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	password := "password123"
	_, err := service.AuthSetup(ctx, &AuthSetupRequest{
		Password: password,
	})
	require.NoError(t, err)

	testImportKey(t, service, alice)

	_, err = service.Sign(context.TODO(), &SignRequest{Data: []byte("test"), KID: alice.ID().String()})
	require.NoError(t, err)

	_, err = service.AuthLock(ctx, &AuthLockRequest{})
	require.NoError(t, err)
	require.Empty(t, service.auth.tokens)

	_, err = service.Sign(context.TODO(), &SignRequest{Data: []byte("test"), KID: alice.ID().String()})
	require.EqualError(t, err, "keyring is locked")
}

func TestAuthSetup(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	_, err := service.AuthSetup(ctx, &AuthSetupRequest{Password: "short"})
	require.EqualError(t, err, "password too short")

	setupResp, err := service.AuthSetup(ctx, &AuthSetupRequest{Password: "password123"})
	require.NoError(t, err)
	require.NotEmpty(t, setupResp.AuthToken)
}
