package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	// server.SetContextLogger(server.NewContextLogger(server.DebugLevel))

	env := newEnv(t)
	srv := newTestServer(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	// GET /msgs/:kid/:rid
	req, err := api.NewRequest("GET", keys.Path("msgs", alice.ID(), charlie.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"messages not found"}}`, body)

	// PUT /msgs/:kid/:rid/:id (no body)
	req, err = api.NewRequest("PUT", keys.Path("msgs", alice.ID(), charlie.ID(), keys.Rand3262()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected := `{"error":{"code":400,"message":"missing body"}}`
	require.Equal(t, expected, body)

	// PUT /msgs/:kid/:rid/:id
	req, err = api.NewRequest("PUT", keys.Path("msgs", alice.ID(), charlie.ID(), keys.Rand3262()), bytes.NewReader([]byte("test1")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// POST /msgs/:kid/:rid/:id (invalid method)
	req, err = api.NewRequest("POST", keys.Path("msgs", alice.ID(), charlie.ID(), keys.Rand3262()), bytes.NewReader([]byte{}), clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusMethodNotAllowed, code)
	require.Equal(t, `{"error":{"code":405,"message":"method not allowed"}}`, body)

	// GET /msgs/:kid/:rid (alice)
	req, err = api.NewRequest("GET", keys.Path("msgs", alice.ID(), charlie.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	require.Equal(t, "1234567890012", resp.Version)
	require.Equal(t, 1, len(resp.Messages))
	require.NotEmpty(t, resp.Messages[0].ID)
	require.Equal(t, []byte("test1"), resp.Messages[0].Data)

	// GET /msgs/:kid/:rid (charlie)
	req, err = api.NewRequest("GET", keys.Path("msgs", charlie.ID(), alice.ID()), nil, clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	t.Logf("body: %s", body)
	require.Equal(t, http.StatusOK, code)
	var charlieResp api.MessagesResponse
	err = json.Unmarshal([]byte(body), &charlieResp)
	require.NoError(t, err)
	require.Equal(t, "1234567890014", charlieResp.Version)
	require.Equal(t, 1, len(charlieResp.Messages))
	require.NotEmpty(t, charlieResp.Messages[0].ID)
	require.Equal(t, []byte("test1"), charlieResp.Messages[0].Data)

	// GET /msgs/:kid/:rid?version=1234567890015
	req, err = api.NewRequest("GET", keys.Path("msgs", alice.ID(), charlie.ID())+"?version=1234567890015", nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp2 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp2)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp2.Messages))
	require.Equal(t, "1234567890015", resp2.Version)

	// PUT /msgs/:kid/:rid/:id
	req, err = api.NewRequest("PUT", keys.Path("msgs", alice.ID(), charlie.ID(), keys.Rand3262()), bytes.NewReader([]byte("test2")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	req, err = api.NewRequest("PUT", keys.Path("msgs", alice.ID(), charlie.ID(), keys.Rand3262()), bytes.NewReader([]byte("test3")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// GET /msgs/:kid/:rid (alice)
	req, err = api.NewRequest("GET", keys.Path("msgs", alice.ID(), charlie.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp3 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp3)
	require.NoError(t, err)
	require.Equal(t, 3, len(resp3.Messages))
	require.Equal(t, []byte("test1"), resp3.Messages[0].Data)
	require.Equal(t, []byte("test2"), resp3.Messages[1].Data)
	require.Equal(t, []byte("test3"), resp3.Messages[2].Data)

	// GET /msgs/:kid/:rid (charlie)
	req, err = api.NewRequest("GET", keys.Path("msgs", charlie.ID(), alice.ID()), nil, clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var charlieResp2 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &charlieResp2)
	require.NoError(t, err)
	require.Equal(t, 3, len(charlieResp2.Messages))
	require.Equal(t, []byte("test1"), charlieResp2.Messages[0].Data)
	require.Equal(t, []byte("test2"), charlieResp2.Messages[1].Data)
	require.Equal(t, []byte("test3"), charlieResp2.Messages[2].Data)

	// GET /msgs/:kid/:rid (descending, limit=2)
	req, err = api.NewRequest("GET", keys.Path("msgs", alice.ID(), charlie.ID())+"?direction=desc&limit=2", nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp4 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp4)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp4.Messages))
	require.Equal(t, []byte("test3"), resp4.Messages[0].Data)
	require.Equal(t, []byte("test2"), resp4.Messages[1].Data)

	// PUT /msgs/:kid/:rid/:id (self)
	id := keys.Rand3262()
	req, err = api.NewRequest("PUT", keys.Path("msgs", alice.ID(), alice.ID(), id), bytes.NewReader([]byte("hi")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// TODO: Get non-expiring message
	// GET /msgs/:kid/:rid/:id (self)
	// req, err = api.NewRequest("GET", keys.Path("msgs", alice.ID(), alice.ID(), id), nil, clock.Now(), alice)
	// require.NoError(t, err)
	// code, _, _ = srv.Serve(req)
	// require.Equal(t, http.StatusOK, code)
	// require.Equal(t, `{}`, body)

	// GET /msgs/:kid/:rid (charlie, invalid)
	req, err = api.NewRequest("GET", keys.Path("msgs", charlie.ID(), alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)

	// PUT /msgs/:kid/:rid/:id (message too large)
	large := bytes.Repeat([]byte{0x01}, 513*1024)
	req, err = api.NewRequest("PUT", keys.Path("msgs", alice.ID(), charlie.ID(), keys.Rand3262()), bytes.NewReader(large), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"message too large (greater than 512KiB)"}}`, body)
}

func TestMessagesAuth(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	env := newEnv(t)
	srv := newTestServer(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	// GET /msgs/:kid/:rid (no auth)
	req, err := http.NewRequest("GET", keys.Path("msgs", alice.ID(), charlie.ID()), nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusUnauthorized, code)
	require.Equal(t, `{"error":{"code":401,"message":"missing Authorization header"}}`, body)

	// GET /msgs/:kid/:rid
	req, err = api.NewRequest("GET", keys.Path("msgs", alice.ID(), charlie.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"messages not found"}}`, body)

	// Replay last request
	reqReplay, err := http.NewRequest("GET", req.URL.String(), nil)
	reqReplay.Header.Set("Authorization", req.Header.Get("Authorization"))
	require.NoError(t, err)
	code, _, body = srv.Serve(reqReplay)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"nonce collision"}}`, body)

	// GET /msgs/:kid/:rid (invalid authorization)
	authHeader := req.Header.Get("Authorization")
	randKey := keys.GenerateEdX25519Key()
	sig := strings.Split(authHeader, ":")[1]
	req, err = api.NewRequest("GET", keys.Path("msgs", randKey.ID(), charlie.ID()), nil, clock.Now(), randKey)
	require.NoError(t, err)
	req.Header.Set("Authorization", randKey.ID().String()+":"+sig)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"verify failed"}}`, body)

	// PUT /msgs/:kid/:rid/:id (invalid recipient)
	req, err = api.NewRequest("PUT", keys.Path("msgs", bob.ID(), charlie.ID(), keys.Rand3262()), bytes.NewReader([]byte("hi")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)
}

func TestMessageExpiring(t *testing.T) {
	// api.SetLogger(api.NewLogger(api.DebugLevel))

	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	// PUT /msgs/:kid/:rid/:id (alice to charlie, 15m)
	req, err := api.NewRequest("PUT", keys.Path("msgs", alice.ID(), charlie.ID(), "wormhole")+"?expire=15m", bytes.NewReader([]byte("hi")), env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /msgs/:kid/:rid/:id (charlie from alice)
	req, err = api.NewRequest("GET", keys.Path("msgs", charlie.ID(), alice.ID(), "wormhole"), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "hi", body)

	// GET again
	req, err = api.NewRequest("GET", keys.Path("msgs", charlie.ID(), alice.ID(), "wormhole"), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /msgs/:kid/:rid/:id (alice to charlie, 15m)
	req, err = api.NewRequest("PUT", keys.Path("msgs", alice.ID(), charlie.ID(), "wormhole")+"?expire=15m", bytes.NewReader([]byte("hi")), env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// DEL (invalid auth)
	req, err = api.NewRequest("DELETE", keys.Path("msgs", alice.ID(), charlie.ID(), "wormhole"), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// DEL /msgs/:kid/:rid/:d
	req, err = api.NewRequest("DELETE", keys.Path("msgs", alice.ID(), charlie.ID(), "wormhole"), nil, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET (after delete)
	req, err = api.NewRequest("GET", keys.Path("msgs", charlie.ID(), alice.ID(), "wormhole"), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /msgs/:kid/:rid/:id (expire 1s)
	req, err = api.NewRequest("PUT", keys.Path("msgs", alice.ID(), charlie.ID(), "wormhole")+"?expire=1ms", bytes.NewReader([]byte("hi")), env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)
	time.Sleep(time.Millisecond)

	// GET (after expire)
	req, err = api.NewRequest("GET", keys.Path("msgs", charlie.ID(), alice.ID(), "wormhole"), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)
}
