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
	// SetContextLogger(NewContextLogger(DebugLevel))
	// firestore.SetContextLogger(NewContextLogger(DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	users := testUserStore(t, fi, rq, clock)
	srv := newTestServer(t, clock, fi, users)

	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	// GET /messages/:kid
	req, err := api.NewRequest("GET", keys.Path("messages", charlie.ID()), nil, clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"messages not found"}}`, body)

	// POST /messages/:kid (no body)
	req, err = api.NewRequest("POST", keys.Path("messages", charlie.ID()), nil, clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected := `{"error":{"code":400,"message":"missing body"}}`
	require.Equal(t, expected, body)

	// POST /messages/:kid
	req, err = api.NewRequest("POST", keys.Path("messages", charlie.ID()), bytes.NewReader([]byte("hi")), clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var msgResp api.MessageResponse
	err = json.Unmarshal([]byte(body), &msgResp)
	require.NoError(t, err)
	require.NotEmpty(t, msgResp.ID)

	// PUT /messages/:kid (invalid method)
	req, err = api.NewRequest("PUT", keys.Path("messages", charlie.ID()), bytes.NewReader([]byte{}), clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusMethodNotAllowed, code)
	require.Equal(t, `{"error":{"code":405,"message":"method not allowed"}}`, body)

	// GET /messages/:kid
	req, err = api.NewRequest("GET", keys.Path("messages", charlie.ID()), nil, clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	require.Equal(t, keys.ID("kex1a4yj333g68pvd6hfqvufqkv4vy54jfe6t33ljd3kc9rpfty8xlgs2u3qxr"), resp.KID)
	require.Equal(t, "1234567890011", resp.Version)
	require.Equal(t, 1, len(resp.Messages))
	require.NotEmpty(t, resp.Messages[0].ID)
	require.Equal(t, []byte("hi"), resp.Messages[0].Data)

	// GET /messages/:kid?version=1234567890012
	req, err = api.NewRequest("GET", keys.Path("messages", charlie.ID())+"?version=1234567890012", nil, clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp2 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp2)
	require.NoError(t, err)
	require.Equal(t, keys.ID("kex1a4yj333g68pvd6hfqvufqkv4vy54jfe6t33ljd3kc9rpfty8xlgs2u3qxr"), resp2.KID)
	require.Equal(t, "1234567890012", resp2.Version)
	require.Equal(t, 0, len(resp2.Messages))
}

func TestMessagesAuth(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	users := testUserStore(t, fi, rq, clock)
	srv := newTestServer(t, clock, fi, users)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// GET /messages/:kid (no auth)
	req, err := http.NewRequest("GET", keys.Path("messages", alice.ID()), nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusUnauthorized, code)
	require.Equal(t, `{"error":{"code":401,"message":"missing Authorization header"}}`, body)

	// GET /messages/:kid
	req, err = api.NewRequest("GET", keys.Path("messages", alice.ID()), nil, clock.Now(), alice)
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

	// GET /messages/:kid (invalid authorization)
	authHeader := req.Header.Get("Authorization")
	randKey := keys.GenerateEdX25519Key()
	sig := strings.Split(authHeader, ":")[1]
	req, err = api.NewRequest("GET", keys.Path("messages", randKey.ID()), nil, clock.Now(), randKey)
	require.NoError(t, err)
	req.Header.Set("Authorization", randKey.ID().String()+":"+sig)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"verify failed"}}`, body)

	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	// POST /messages/:kid (invalid recipient)
	req, err = api.NewRequest("POST", keys.Path("messages", bob.ID()), bytes.NewReader([]byte("hi")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)
}
