package service

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuth(t *testing.T) {
	cfg, closeFn := testConfig(t)
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
	cfg, closeFn := testConfig(t)
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
	clock := newClock()
	fi := testFire(t, clock)
	service, closeFn := testServiceFire(t, fi, clock)
	defer closeFn()
	ctx := context.TODO()

	pepper, err := keys.BytesToPhrase(keys.Rand32()[:])
	require.NoError(t, err)
	setupResp, err := service.AuthSetup(ctx, &AuthSetupRequest{Password: "password123", Pepper: pepper})
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
	clock := newClock()
	fi := testFire(t, clock)
	service, closeFn := testServiceFire(t, fi, clock)
	defer closeFn()
	ctx := context.TODO()

	b := bytes.Repeat([]byte{0x01}, 32)
	pepper, err := keys.BytesToPhrase(b)
	require.NoError(t, err)
	setupResp, err := service.AuthSetup(ctx, &AuthSetupRequest{Password: "short", Pepper: pepper, PublishPublicKey: true})
	require.EqualError(t, err, "password too short")
	setupResp, err = service.AuthSetup(ctx, &AuthSetupRequest{Password: "password123", Pepper: "invalid pepper", PublishPublicKey: true})
	require.EqualError(t, err, "invalid pepper: invalid phrase")
	setupResp, err = service.AuthSetup(ctx, &AuthSetupRequest{Password: "password123", Pepper: pepper, PublishPublicKey: true})
	require.NoError(t, err)
	kid := setupResp.KID
	require.Equal(t, "DP1ALkPtRQrCEmXrEDFKbL1S5b5TiGdU5hCTVfwwVcaj", kid)

	keysResp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(keysResp.Keys))
	require.Equal(t, "DP1ALkPtRQrCEmXrEDFKbL1S5b5TiGdU5hCTVfwwVcaj", keysResp.Keys[0].KID)

	// New service to recover
	recover, closeFn := testServiceFire(t, fi, clock)
	defer closeFn()

	_, err = recover.AuthSetup(ctx, &AuthSetupRequest{
		Password: "password123",
		Pepper:   pepper,
	})
	st, _ := status.FromError(err)
	require.NotNil(t, st)
	require.Equal(t, st.Code(), codes.AlreadyExists)
	require.Equal(t, st.Message(), "key already exists (use recover instead)")

	recoverResp, err := recover.AuthSetup(ctx, &AuthSetupRequest{
		Password: "password123",
		Pepper:   pepper,
		Recover:  true,
	})
	require.NoError(t, err)
	require.Equal(t, "DP1ALkPtRQrCEmXrEDFKbL1S5b5TiGdU5hCTVfwwVcaj", recoverResp.KID)

	keysResp, err = recover.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(keysResp.Keys))
	require.Equal(t, "DP1ALkPtRQrCEmXrEDFKbL1S5b5TiGdU5hCTVfwwVcaj", keysResp.Keys[0].KID)

	_, err = recover.AuthSetup(ctx, &AuthSetupRequest{
		Password: "password1234",
		Pepper:   pepper,
		Recover:  true,
	})
	st, _ = status.FromError(err)
	require.NotNil(t, st)
	require.Equal(t, st.Code(), codes.NotFound)
	require.Equal(t, st.Message(), "key not found or wasn't published (use -force to bypass)")
}

func TestAuthRecoverNoPublish(t *testing.T) {
	clock := newClock()
	fi := testFire(t, clock)
	service, closeFn := testServiceFire(t, fi, clock)
	defer closeFn()
	ctx := context.TODO()

	b := bytes.Repeat([]byte{0x01}, 32)
	pepper, err := keys.BytesToPhrase(b)
	require.NoError(t, err)
	setupResp, err := service.AuthSetup(ctx, &AuthSetupRequest{Password: "password123", Pepper: pepper})
	require.NoError(t, err)
	kid := setupResp.KID
	require.Equal(t, "DP1ALkPtRQrCEmXrEDFKbL1S5b5TiGdU5hCTVfwwVcaj", kid)

	keysResp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(keysResp.Keys))
	require.Equal(t, "DP1ALkPtRQrCEmXrEDFKbL1S5b5TiGdU5hCTVfwwVcaj", keysResp.Keys[0].KID)

	// New service to recover
	recover, closeFn := testServiceFire(t, fi, clock)
	defer closeFn()

	_, err = recover.AuthSetup(ctx, &AuthSetupRequest{
		Password: "password123",
		Pepper:   pepper,
		Recover:  true,
	})
	st, _ := status.FromError(err)
	require.NotNil(t, st)
	require.Equal(t, st.Code(), codes.NotFound)
	require.Equal(t, st.Message(), "key not found or wasn't published (use -force to bypass)")

	recoverResp, err := recover.AuthSetup(ctx, &AuthSetupRequest{
		Password: "password123",
		Pepper:   pepper,
		Recover:  true,
		Force:    true,
	})
	require.Equal(t, "DP1ALkPtRQrCEmXrEDFKbL1S5b5TiGdU5hCTVfwwVcaj", recoverResp.KID)
}
