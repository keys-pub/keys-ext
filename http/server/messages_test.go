package server_test

import (
	"bytes"
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

	group := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x04}, 32)))

	// GET /messages/:kid
	req, err := api.NewRequest("GET", keys.Path("messages", group.ID()), nil, clock.Now(), group)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"messages not found"}}`, body)

	// PUT /messages/:kid/:id (no body)
	req, err = api.NewRequest("PUT", keys.Path("messages", group.ID(), keys.RandString(32)), nil, clock.Now(), group)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected := `{"error":{"code":400,"message":"missing body"}}`
	require.Equal(t, expected, body)

	// PUT /messages/:kid/:id
	id := "H1zXH53Xt3JJGx51ruhqk1p83q3VFGmUQCunR51fAsSu"
	req, err = api.NewRequest("PUT", keys.Path("messages", group.ID(), id), bytes.NewReader([]byte("hi")), clock.Now(), group)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// POST /messages/:kid/:id (invalid method)
	req, err = api.NewRequest("POST", keys.Path("messages", group.ID(), keys.RandString(32)), bytes.NewReader([]byte{}), clock.Now(), group)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusMethodNotAllowed, code)
	require.Equal(t, `{"error":{"code":405,"message":"method not allowed"}}`, body)

	// GET /messages/:kid
	req, err = api.NewRequest("GET", keys.Path("messages", group.ID()), nil, clock.Now(), group)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedMessages := `{"kid":"kex1e2f6c9c9rpc8r4nms0rl7rh7syyw3mz9xpt46aexs7fn8k76he7qsa5489","messages":[{"data":"aGk=","id":"H1zXH53Xt3JJGx51ruhqk1p83q3VFGmUQCunR51fAsSu","path":"/messages/kex1e2f6c9c9rpc8r4nms0rl7rh7syyw3mz9xpt46aexs7fn8k76he7qsa5489-H1zXH53Xt3JJGx51ruhqk1p83q3VFGmUQCunR51fAsSu"}],"version":"1234567890011"}`
	require.Equal(t, expectedMessages, body)

	// GET /messages/:kid?version=1234567890012
	req, err = api.NewRequest("GET", keys.Path("messages", group.ID())+"?version=1234567890012", nil, clock.Now(), group)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedMessages = `{"kid":"kex1e2f6c9c9rpc8r4nms0rl7rh7syyw3mz9xpt46aexs7fn8k76he7qsa5489","messages":[],"version":"1234567890012"}`
	require.Equal(t, expectedMessages, body)
}

func TestMessagesAuth(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	users := testUserStore(t, fi, rq, clock)
	srv := newTestServer(t, clock, fi, users)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// GET /messages/:id (no auth)
	req, err := http.NewRequest("GET", keys.Path("messages", keys.RandString(32)), nil)
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
}
