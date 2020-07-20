package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/docs"
	"github.com/stretchr/testify/require"
)

func TestVault(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	testVault(t, env, alice)
}

func testVault(t *testing.T, env *env, alice *keys.EdX25519Key) {
	srv := newTestServer(t, env)
	clock := env.clock

	rand := keys.GenerateEdX25519Key()

	// GET /vault/:kid (not found)
	req, err := api.NewRequest("GET", docs.Path("vault", alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault not found"}}`, body)

	// HEAD /vault/:kid (not found)
	req, err = api.NewRequest("HEAD", docs.Path("vault", alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)

	// POST /vault/:kid
	req, err = api.NewRequest("POST", docs.Path("vault", alice.ID()), bytes.NewReader([]byte("test1")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /vault/:kid
	req, err = api.NewRequest("GET", docs.Path("vault", alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.VaultResponse
	err = json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Index)
	require.Equal(t, 1, len(resp.Vault))
	require.Equal(t, []byte("test1"), resp.Vault[0].Data)

	// HEAD /vault/:kid
	req, err = api.NewRequest("HEAD", docs.Path("vault", alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// GET /vault/:kid?idx=next
	req, err = api.NewRequest("GET", docs.Path("vault", alice.ID())+"?idx="+strconv.Itoa(int(resp.Index)), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp2 api.VaultResponse
	err = json.Unmarshal([]byte(body), &resp2)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp2.Vault))
	require.Equal(t, resp.Index, resp2.Index)

	// POST /vault/:kid
	req, err = api.NewRequest("POST", docs.Path("vault", alice.ID()), bytes.NewReader([]byte("test2")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)
	// POST /vault/:kid
	req, err = api.NewRequest("POST", docs.Path("vault", alice.ID()), bytes.NewReader([]byte("test3")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /vault/:kid?idx=next
	req, err = api.NewRequest("GET", docs.Path("vault", alice.ID())+"?idx="+strconv.Itoa(int(resp.Index)), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp3 api.VaultResponse
	err = json.Unmarshal([]byte(body), &resp3)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp3.Vault))
	require.Equal(t, []byte("test2"), resp3.Vault[0].Data)
	require.Equal(t, []byte("test3"), resp3.Vault[1].Data)

	// PUT /vault/:kid
	vault := []*api.Data{
		&api.Data{Data: []byte("test4")},
		&api.Data{Data: []byte("test5")},
		&api.Data{Data: []byte("test6")},
		&api.Data{Data: []byte("test7")},
		&api.Data{Data: []byte("test8")},
		&api.Data{Data: []byte("test9")},
	}
	data, err := json.Marshal(vault)
	require.NoError(t, err)
	req, err = api.NewRequest("PUT", docs.Path("vault", alice.ID()), bytes.NewReader(data), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /vault/:kid?idx=next
	req, err = api.NewRequest("GET", docs.Path("vault", alice.ID())+"?idx="+strconv.Itoa(int(resp3.Index)), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp4 api.VaultResponse
	err = json.Unmarshal([]byte(body), &resp4)
	require.NoError(t, err)
	require.Equal(t, 6, len(resp4.Vault))
	require.Equal(t, []byte("test4"), resp4.Vault[0].Data)
	require.Equal(t, []byte("test5"), resp4.Vault[1].Data)
	require.Equal(t, []byte("test6"), resp4.Vault[2].Data)
	require.Equal(t, []byte("test7"), resp4.Vault[3].Data)
	require.Equal(t, []byte("test8"), resp4.Vault[4].Data)
	require.Equal(t, []byte("test9"), resp4.Vault[5].Data)

	// DEL (invalid auth)
	req, err = api.NewRequest("DELETE", docs.Path("vault", alice.ID()), nil, env.clock.Now(), rand)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// DEL /vault/:kid
	req, err = api.NewRequest("DELETE", docs.Path("vault", alice.ID()), nil, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// DEL /vault/:kid (again)
	req, err = api.NewRequest("DELETE", docs.Path("vault", alice.ID()), nil, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault was deleted"}}`, body)

	// GET /vault/:kid
	req, err = api.NewRequest("GET", docs.Path("vault", alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault was deleted"}}`, body)

	// HEAD /vault/:kid
	req, err = api.NewRequest("HEAD", docs.Path("vault", alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault was deleted"}}`, body)

	// PUT /vault/:kid (deleted)
	vault = []*api.Data{
		&api.Data{Data: []byte("testdeleted")},
	}
	data, err = json.Marshal(vault)
	require.NoError(t, err)
	req, err = api.NewRequest("PUT", docs.Path("vault", alice.ID()), bytes.NewReader(data), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault was deleted"}}`, body)
}

func TestVaultAuth(t *testing.T) {
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))

	testVaultAuth(t, env, alice)
}

func testVaultAuth(t *testing.T, env *env, alice *keys.EdX25519Key) {
	srv := newTestServer(t, env)
	clock := env.clock

	randKey := keys.GenerateEdX25519Key()

	// GET /vault/:kid (no auth)
	req, err := http.NewRequest("GET", docs.Path("vault", alice.ID()), nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusUnauthorized, code)
	require.Equal(t, `{"error":{"code":401,"message":"missing Authorization header"}}`, body)

	// GET /vault/:kid (invalid key)
	req, err = api.NewRequest("GET", docs.Path("vault", alice.ID()), nil, clock.Now(), randKey)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// POST /vault/:kid/id1/1 (invalid key)
	req, err = api.NewRequest("POST", docs.Path("vault", alice.ID()), bytes.NewReader([]byte("test")), clock.Now(), randKey)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// GET /vault/:kid
	req, err = api.NewRequest("GET", docs.Path("vault", alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault not found"}}`, body)

	// Replay last request
	reqReplay, err := http.NewRequest("GET", req.URL.String(), nil)
	reqReplay.Header.Set("Authorization", req.Header.Get("Authorization"))
	require.NoError(t, err)
	code, _, body = srv.Serve(reqReplay)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"nonce collision"}}`, body)

	// GET /vault/:kid (invalid authorization)
	authHeader := req.Header.Get("Authorization")
	sig := strings.Split(authHeader, ":")[1]
	req, err = api.NewRequest("GET", docs.Path("vault", alice.ID()), nil, clock.Now(), randKey)
	require.NoError(t, err)
	req.Header.Set("Authorization", randKey.ID().String()+":"+sig)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"verify failed"}}`, body)
}
