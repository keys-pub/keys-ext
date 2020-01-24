package service

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	seed := bytes.Repeat([]byte{0x01}, 32)
	keyBackup := seedToSaltpack(password, seed)
	setupResp, err := service.AuthSetup(ctx, &AuthSetupRequest{
		Password:  password,
		KeyBackup: keyBackup,
	})
	require.NoError(t, err)
	kid := setupResp.KID

	_, err = service.Sign(context.TODO(), &SignRequest{Data: []byte("test"), KID: kid})
	require.NoError(t, err)

	_, err = service.AuthLock(ctx, &AuthLockRequest{})
	require.NoError(t, err)
	require.Empty(t, service.auth.tokens)

	_, err = service.Sign(context.TODO(), &SignRequest{Data: []byte("test"), KID: kid})
	require.EqualError(t, err, "keyring is locked")
}

func TestAuthSetup(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	genResp, err := service.AuthGenerate(ctx, &AuthGenerateRequest{Password: "password123"})
	require.NoError(t, err)
	require.NotEmpty(t, genResp.KeyBackup)

	seed := bytes.Repeat([]byte{0x01}, 32)
	setupResp, err := service.AuthSetup(ctx, &AuthSetupRequest{Password: "short", KeyBackup: seedToSaltpack("short", seed)})
	require.EqualError(t, err, "password too short")

	keyBackup := seedToSaltpack("password123", seed)
	setupResp, err = service.AuthSetup(ctx, &AuthSetupRequest{Password: "password123", KeyBackup: "invalid recovery"})
	st, _ := status.FromError(err)
	require.NotNil(t, st)
	require.Equal(t, codes.PermissionDenied, st.Code())
	require.Equal(t, "invalid key backup: failed to parse saltpack: missing saltpack start", st.Message())

	setupResp, err = service.AuthSetup(ctx, &AuthSetupRequest{Password: "password123", KeyBackup: keyBackup})
	require.NoError(t, err)
	kid := setupResp.KID
	require.Equal(t, "kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen", kid)

	itemsResp, err := service.Items(ctx, &ItemsRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(itemsResp.Items))
	require.Equal(t, "kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen", itemsResp.Items[0].ID)
}

func TestAuthRecover(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	password := "password123"
	seed := bytes.Repeat([]byte{0x01}, 32)
	keyBackup := seedToSaltpack(password, seed)

	// Invalid password
	_, err := service.AuthSetup(ctx, &AuthSetupRequest{
		Password:  "password1234",
		KeyBackup: keyBackup,
	})
	st, _ := status.FromError(err)
	require.NotNil(t, st)
	require.Equal(t, codes.PermissionDenied, st.Code())
	require.Equal(t, "invalid key backup: failed to decrypt with a password: secretbox open failed", st.Message())

	// Valid recovery
	recoverResp, err := service.AuthSetup(ctx, &AuthSetupRequest{
		Password:  "password123",
		KeyBackup: keyBackup,
	})
	require.NoError(t, err)
	require.Equal(t, "kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen", recoverResp.KID)
}
