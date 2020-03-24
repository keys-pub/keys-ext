package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

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

	// GET /messages/:kid/:rid
	req, err := api.NewRequest("GET", keys.Path("messages", alice.ID(), charlie.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"messages not found"}}`, body)

	// POST /messages/:kid/:rid (no body)
	req, err = api.NewRequest("POST", keys.Path("messages", alice.ID(), charlie.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected := `{"error":{"code":400,"message":"missing body"}}`
	require.Equal(t, expected, body)

	// POST /messages/:kid/:rid
	req, err = api.NewRequest("POST", keys.Path("messages", alice.ID(), charlie.ID()), bytes.NewReader([]byte("hi")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	t.Logf("body: %s", body)
	require.Equal(t, http.StatusOK, code)
	var msgResp api.MessageResponse
	err = json.Unmarshal([]byte(body), &msgResp)
	require.NoError(t, err)
	require.NotEmpty(t, msgResp.ID)

	// PUT /messages/:kid/:rid (invalid method)
	req, err = api.NewRequest("PUT", keys.Path("messages", alice.ID(), charlie.ID()), bytes.NewReader([]byte{}), clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusMethodNotAllowed, code)
	require.Equal(t, `{"error":{"code":405,"message":"method not allowed"}}`, body)

	// GET /messages/:kid/:rid
	req, err = api.NewRequest("GET", keys.Path("messages", alice.ID(), charlie.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	require.Equal(t, "1234567890012", resp.Version)
	require.Equal(t, 1, len(resp.Messages))
	require.NotEmpty(t, resp.Messages[0].ID)
	require.Equal(t, []byte("hi"), resp.Messages[0].Data)

	// GET /messages/:kid/:rid (charlie)
	req, err = api.NewRequest("GET", keys.Path("messages", charlie.ID(), alice.ID()), nil, clock.Now(), charlie)
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
	require.Equal(t, []byte("hi"), charlieResp.Messages[0].Data)

	// GET /messages/:kid/:rid?version=1234567890015
	req, err = api.NewRequest("GET", keys.Path("messages", alice.ID(), charlie.ID())+"?version=1234567890015", nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp2 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp2)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp2.Messages))
	require.Equal(t, "1234567890015", resp2.Version)

	// POST /messages/:kid/:rid (channel=wormhole)
	req, err = api.NewRequest("POST", keys.Path("messages", alice.ID(), charlie.ID())+"?channel=wormhole", bytes.NewReader([]byte("test1")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	req, err = api.NewRequest("POST", keys.Path("messages", alice.ID(), charlie.ID())+"?channel=wormhole", bytes.NewReader([]byte("test2")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	req, err = api.NewRequest("POST", keys.Path("messages", alice.ID(), charlie.ID())+"?channel=wormhole", bytes.NewReader([]byte("test3")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// GET /messages/:kid/:rid (channel=wormhole)
	req, err = api.NewRequest("GET", keys.Path("messages", alice.ID(), charlie.ID())+"?channel=wormhole&dir=asc", nil, clock.Now(), alice)
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

	// GET /messages/:kid/:rid (charlie, channel=wormhole)
	req, err = api.NewRequest("GET", keys.Path("messages", charlie.ID(), alice.ID())+"?channel=wormhole&dir=asc", nil, clock.Now(), charlie)
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

	// GET /messages/:kid/:rid (channel=wormhole, descending, limit=2)
	req, err = api.NewRequest("GET", keys.Path("messages", alice.ID(), charlie.ID())+"?channel=wormhole&direction=desc&limit=2", nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp4 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp4)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp4.Messages))
	require.Equal(t, []byte("test3"), resp4.Messages[0].Data)
	require.Equal(t, []byte("test2"), resp4.Messages[1].Data)

	// POST /messages/:kid/:rid (self)
	req, err = api.NewRequest("POST", keys.Path("messages", alice.ID(), alice.ID()), bytes.NewReader([]byte("hi")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var msgResp2 api.MessageResponse
	err = json.Unmarshal([]byte(body), &msgResp2)
	require.NoError(t, err)
	require.NotEmpty(t, msgResp2.ID)

	// GET /messages/:kid/:rid (charlie, invalid)
	req, err = api.NewRequest("GET", keys.Path("messages", charlie.ID(), alice.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)

	// POST /messages/:kid/:rid (channel invalid)
	req, err = api.NewRequest("POST", keys.Path("messages", alice.ID(), charlie.ID())+"?channel=channelnametoolongtoolongtoolong", bytes.NewReader([]byte("test1")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"channel name too long"}}`, body)

	// GET /messages/:kid/:rid (channel invalid)
	req, err = api.NewRequest("GET", keys.Path("messages", alice.ID(), charlie.ID()+"?channel=channelnametoolongtoolongtoolong"), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"channel name too long"}}`, body)

	// POST /messages/:kid/:rid (message too large)
	large := bytes.Repeat([]byte{0x01}, 513*1024)
	req, err = api.NewRequest("POST", keys.Path("messages", alice.ID(), charlie.ID()), bytes.NewReader(large), clock.Now(), alice)
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

	// GET /messages/:kid/:rid (no auth)
	req, err := http.NewRequest("GET", keys.Path("messages", alice.ID(), charlie.ID()), nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusUnauthorized, code)
	require.Equal(t, `{"error":{"code":401,"message":"missing Authorization header"}}`, body)

	// GET /messages/:kid/:rid
	req, err = api.NewRequest("GET", keys.Path("messages", alice.ID(), charlie.ID()), nil, clock.Now(), alice)
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

	// GET /messages/:kid/:rid (invalid authorization)
	authHeader := req.Header.Get("Authorization")
	randKey := keys.GenerateEdX25519Key()
	sig := strings.Split(authHeader, ":")[1]
	req, err = api.NewRequest("GET", keys.Path("messages", randKey.ID(), charlie.ID()), nil, clock.Now(), randKey)
	require.NoError(t, err)
	req.Header.Set("Authorization", randKey.ID().String()+":"+sig)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"verify failed"}}`, body)

	// POST /messages/:kid/:rid (invalid recipient)
	req, err = api.NewRequest("POST", keys.Path("messages", bob.ID(), charlie.ID()), bytes.NewReader([]byte("hi")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)
}
