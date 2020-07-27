package service

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestKey(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	// db.SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetupGithub(t, env, service, alice, "alice")

	// Alice
	resp, err := service.Key(ctx, &KeyRequest{
		Identity: alice.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, alice.ID().String(), resp.Key.ID)

	// Alice (user)
	resp, err = service.Key(ctx, &KeyRequest{
		Identity: "alice@github",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Key)
	require.Equal(t, alice.ID().String(), resp.Key.ID)

	// TODO: Test update
}

func TestFmtKey(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	ak, err := service.keyToRPC(ctx, alice)
	require.NoError(t, err)
	var buf bytes.Buffer
	fmtKey(&buf, ak)
	require.Equal(t, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077\n", buf.String())

	testUserSetupGithub(t, env, service, alice, "alice")

	ak2, err := service.keyToRPC(ctx, alice)
	require.NoError(t, err)

	var buf2 bytes.Buffer
	fmtKey(&buf2, ak2)
	require.Equal(t, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 alice@github\n", buf2.String())
}

func TestFmtVerifiedEncrypt(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	ak, err := service.keyToRPC(ctx, alice)
	require.NoError(t, err)
	var buf bytes.Buffer
	fmtVerifiedEncrypt(&buf, ak, SaltpackEncrypt)
	require.Equal(t, "verified: kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 saltpack-encrypt\n", buf.String())

	testUserSetupGithub(t, env, service, alice, "alice")

	ak2, err := service.keyToRPC(ctx, alice)
	require.NoError(t, err)

	var buf2 bytes.Buffer
	fmtVerifiedEncrypt(&buf2, ak2, SaltpackSigncrypt)
	require.Equal(t, "verified: kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 saltpack-signcrypt alice@github\n", buf2.String())
}

func TestFmtVerified(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	ak, err := service.keyToRPC(ctx, alice)
	require.NoError(t, err)
	var buf bytes.Buffer
	fmtVerified(&buf, ak)
	require.Equal(t, "verified: kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077\n", buf.String())

	testUserSetupGithub(t, env, service, alice, "alice")

	ak2, err := service.keyToRPC(ctx, alice)
	require.NoError(t, err)

	var buf2 bytes.Buffer
	fmtVerified(&buf2, ak2)
	require.Equal(t, "verified: kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 alice@github\n", buf2.String())
}

func TestKeyGenerate(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{Type: EdX25519})
	require.NoError(t, err)

	key, err := service.parseSignKey(genResp.KID, true)
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, key.ID().String(), genResp.KID)
}

func TestKeyRemove(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{Type: EdX25519})
	require.NoError(t, err)
	kid, err := keys.ParseID(genResp.KID)
	require.NoError(t, err)
	key, err := service.vault.EdX25519Key(kid)
	require.NoError(t, err)
	require.NotNil(t, key)

	// Remove
	_, err = service.KeyRemove(ctx, &KeyRemoveRequest{KID: key.ID().String()})
	require.NoError(t, err)

	// Remove (X25519)
	genResp, err = service.KeyGenerate(ctx, &KeyGenerateRequest{Type: X25519})
	require.NoError(t, err)
	_, err = service.KeyRemove(ctx, &KeyRemoveRequest{KID: genResp.KID})
	require.NoError(t, err)

	// Remove (not found)
	randKey := keys.GenerateEdX25519Key()
	_, err = service.KeyRemove(ctx, &KeyRemoveRequest{KID: randKey.ID().String()})
	require.EqualError(t, err, fmt.Sprintf("not found %s", randKey.ID()))
}
