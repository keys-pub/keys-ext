package server_test

import (
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/vault/auth"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

func TestAccountAuth(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServerEnv(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))

	// Auth method
	mk := keys.Rand32()
	password, err := auth.NewPasswordAuth("testpassword", mk)
	require.NoError(t, err)
	sk := keys.Rand32()
	data := secretBoxMarshal(t, password, sk)

	// POST /account/:aid/auths
	req, err := http.NewJSONRequest("POST", dstore.Path("account", alice.ID(), "auths"), &api.AccountAuth{ID: password.ID, Data: data}, http.WithTimestamp(env.clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))

	// GET /account/:aid/auths
	req, err = http.NewAuthRequest("GET", dstore.Path("account", alice.ID(), "auths"), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	resp := api.AccountAuthsResponse{}
	testJSONUnmarshal(t, body, &resp)
	require.Equal(t, http.StatusOK, code)

	var out auth.Auth
	secretBoxUnmarshal(t, resp.Auths[0].Data, &out, sk)
	require.NoError(t, err)
	require.Equal(t, password.ID, out.ID)
	require.Equal(t, password.EncryptedKey, out.EncryptedKey)

}

func secretBoxMarshal(t *testing.T, i interface{}, secretKey *[32]byte) []byte {
	b, err := msgpack.Marshal(i)
	require.NoError(t, err)
	return keys.SecretBoxSeal(b, secretKey)
}

func secretBoxUnmarshal(t *testing.T, b []byte, v interface{}, secretKey *[32]byte) {
	decrypted, err := keys.SecretBoxOpen(b, secretKey)
	require.NoError(t, err)
	err = msgpack.Unmarshal(decrypted, v)
	require.NoError(t, err)
}
