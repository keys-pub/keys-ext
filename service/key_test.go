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
	fmtKey(&buf, ak, "")
	require.Equal(t, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077\n", buf.String())

	testUserSetupGithub(t, env, service, alice, "alice")

	ak2, err := service.keyToRPC(ctx, alice)
	require.NoError(t, err)
	var buf2 bytes.Buffer
	fmtKey(&buf2, ak2, "verified ")
	require.Equal(t, "verified kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 alice@github\n", buf2.String())
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
	key, err := keys.FindEdX25519Key(service.kr, kid)
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

// func TestItem(t *testing.T) {
// 	clock := newClock()
// 	item := keyring.NewItem("account1", []byte("password"), "passphrase", clock.Now())
// 	b, err := msgpack.Marshal(item)
// 	require.NoError(t, err)
// 	expected := `([]uint8) (len=56 cap=64) {
//  00000000  84 a2 69 64 a8 61 63 63  6f 75 6e 74 31 a3 74 79  |..id.account1.ty|
//  00000010  70 aa 70 61 73 73 70 68  72 61 73 65 a3 64 61 74  |p.passphrase.dat|
//  00000020  c4 08 70 61 73 73 77 6f  72 64 a3 63 74 73 d7 ff  |..password.cts..|
//  00000030  00 3d 09 00 49 96 02 d2                           |.=..I...|
// }
// `
// 	require.Equal(t, expected, spew.Sdump(b))
// }
