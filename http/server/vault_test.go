package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/ds"
	"github.com/stretchr/testify/require"
)

func TestVault(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))

	srv := newTestServer(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// GET /vault/:kid (not found)
	req, err := api.NewRequest("GET", ds.Path("vault", alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault not found"}}`, body)

	// POST /vault/:kid/id1
	req, err = api.NewRequest("POST", ds.Path("vault", alice.ID()), bytes.NewReader([]byte("test1.1")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// PUT /vault/:kid (invalid method)
	req, err = api.NewRequest("PUT", ds.Path("vault", alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusMethodNotAllowed, code)
	require.Equal(t, `{"error":{"code":405,"message":"method not allowed"}}`, body)

	// POST /vault/:kid/id1
	req, err = api.NewRequest("POST", ds.Path("vault", alice.ID()), bytes.NewReader([]byte("test1.2")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /vault/:kid
	req, err = api.NewRequest("GET", ds.Path("vault", alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.VaultResponse
	err = json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Version)
	require.Equal(t, 2, len(resp.Items))
	require.Equal(t, []byte("test1.1"), resp.Items[0].Data)
	require.Equal(t, []byte("test1.2"), resp.Items[1].Data)

	// GET /vault/:kid?version=next
	req, err = api.NewRequest("GET", ds.Path("vault", alice.ID())+"?version="+resp.Version, nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp2 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp2)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp2.Messages))
	require.Equal(t, resp.Version, resp2.Version)

	// POST /vault/:kid
	req, err = api.NewRequest("POST", ds.Path("vault", alice.ID()), bytes.NewReader([]byte("test2.1")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)
	// POST /vault/:kid
	req, err = api.NewRequest("POST", ds.Path("vault", alice.ID()), bytes.NewReader([]byte("test3.1")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /vault/:kid?version=next
	req, err = api.NewRequest("GET", ds.Path("vault", alice.ID())+"?version="+resp.Version, nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp3 api.VaultResponse
	err = json.Unmarshal([]byte(body), &resp3)
	require.NoError(t, err)
	require.Equal(t, 3, len(resp3.Items))
	require.Equal(t, []byte("test1.2"), resp3.Items[0].Data)
	require.Equal(t, []byte("test2.1"), resp3.Items[1].Data)
	require.Equal(t, []byte("test3.1"), resp3.Items[2].Data)
}

func TestVaultAuth(t *testing.T) {
	env := newEnv(t)
	srv := newTestServer(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	randKey := keys.GenerateEdX25519Key()

	// GET /vault/:kid (no auth)
	req, err := http.NewRequest("GET", ds.Path("vault", alice.ID()), nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusUnauthorized, code)
	require.Equal(t, `{"error":{"code":401,"message":"missing Authorization header"}}`, body)

	// GET /vault/:kid (invalid key)
	req, err = api.NewRequest("GET", ds.Path("vault", alice.ID()), nil, clock.Now(), randKey)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// POST /vault/:kid/id1/1 (invalid key)
	req, err = api.NewRequest("POST", ds.Path("vault", alice.ID()), bytes.NewReader([]byte("test")), clock.Now(), randKey)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// GET /vault/:kid
	req, err = api.NewRequest("GET", ds.Path("vault", alice.ID()), nil, clock.Now(), alice)
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
	req, err = api.NewRequest("GET", ds.Path("vault", alice.ID()), nil, clock.Now(), randKey)
	require.NoError(t, err)
	req.Header.Set("Authorization", randKey.ID().String()+":"+sig)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"verify failed"}}`, body)
}
