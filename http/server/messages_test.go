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
	"github.com/keys-pub/keys/encoding"
	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	testMessages(t, env, alice, charlie)
}

func testMessages(t *testing.T, env *env, alice *keys.EdX25519Key, charlie *keys.EdX25519Key) {
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))

	srv := newTestServer(t, env)
	clock := env.clock

	// GET /msgs/:kid/:rid
	req, err := api.NewRequest("GET", docs.Path("msgs", alice.ID(), charlie.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"msgs":[],"idx":0}`, body)

	// POST /msgs/:kid/:rid (no body)
	req, err = api.NewRequest("POST", docs.Path("msgs", alice.ID(), charlie.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected := `{"error":{"code":400,"message":"missing body"}}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusBadRequest, code)

	// POST /msgs/:kid/:rid
	content := []byte("test1")
	contentHash := encoding.EncodeBase64(keys.SHA256(content))
	req, err = api.NewRequest("POST", docs.Path("msgs", alice.ID(), charlie.ID()), bytes.NewReader(content), contentHash, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// PUT /msgs/:kid/:rid (invalid method)
	req, err = api.NewRequest("PUT", docs.Path("msgs", alice.ID(), charlie.ID()), bytes.NewReader(content), contentHash, clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusMethodNotAllowed, code)
	require.Equal(t, `{"error":{"code":405,"message":"method not allowed"}}`, body)

	// GET /msgs/:kid/:rid (alice)
	req, err = api.NewRequest("GET", docs.Path("msgs", alice.ID(), charlie.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Index)
	require.Equal(t, 1, len(resp.Messages))
	require.Equal(t, []byte("test1"), resp.Messages[0].Data)

	// GET /msgs/:kid/:rid (charlie)
	req, err = api.NewRequest("GET", docs.Path("msgs", charlie.ID(), alice.ID()), nil, "", clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	// t.Logf("body: %s", body)
	require.Equal(t, http.StatusOK, code)
	var charlieResp api.MessagesResponse
	err = json.Unmarshal([]byte(body), &charlieResp)
	require.NoError(t, err)
	require.Equal(t, int64(1), charlieResp.Index)
	require.Equal(t, 1, len(charlieResp.Messages))
	require.Equal(t, []byte("test1"), charlieResp.Messages[0].Data)

	// GET /msgs/:kid/:rid?idx=next
	req, err = api.NewRequest("GET", docs.Path("msgs", alice.ID(), charlie.ID())+"?idx="+strconv.Itoa(int(charlieResp.Index)), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp2 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp2)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp2.Messages))
	require.Equal(t, charlieResp.Index, resp2.Index)

	// POST /msgs/:kid/:rid
	content2 := []byte("test2")
	contentHash2 := encoding.EncodeBase64(keys.SHA256(content2))
	req, err = api.NewRequest("POST", docs.Path("msgs", alice.ID(), charlie.ID()), bytes.NewReader(content2), contentHash2, clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	content3 := []byte("test3")
	contentHash3 := encoding.EncodeBase64(keys.SHA256(content3))
	req, err = api.NewRequest("POST", docs.Path("msgs", alice.ID(), charlie.ID()), bytes.NewReader(content3), contentHash3, clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// GET /msgs/:kid/:rid (alice)
	req, err = api.NewRequest("GET", docs.Path("msgs", alice.ID(), charlie.ID()), nil, "", clock.Now(), alice)
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
	req, err = api.NewRequest("GET", docs.Path("msgs", charlie.ID(), alice.ID()), nil, "", clock.Now(), charlie)
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
	req, err = api.NewRequest("GET", docs.Path("msgs", alice.ID(), charlie.ID())+"?dir=desc&limit=2", nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp4 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp4)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp4.Messages))
	require.Equal(t, []byte("test3"), resp4.Messages[0].Data)
	require.Equal(t, []byte("test2"), resp4.Messages[1].Data)

	// POST /msgs/:kid/:rid (self)
	content4 := []byte("selfie")
	contentHash4 := encoding.EncodeBase64(keys.SHA256(content4))
	req, err = api.NewRequest("POST", docs.Path("msgs", alice.ID(), alice.ID()), bytes.NewReader(content4), contentHash4, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	t.Logf(body)
	require.Equal(t, http.StatusOK, code)

	// GET /msgs/:kid/:rid (charlie, invalid)
	req, err = api.NewRequest("GET", docs.Path("msgs", charlie.ID(), alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)

	// POST /msgs/:kid/:rid (message too large)
	large := bytes.Repeat([]byte{0x01}, 17*1024)
	largeHash := encoding.EncodeBase64(keys.SHA256(large))
	req, err = api.NewRequest("POST", docs.Path("msgs", alice.ID(), charlie.ID()), bytes.NewReader(large), largeHash, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"message too large (greater than 16KiB)"}}`, body)
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
	req, err := http.NewRequest("GET", docs.Path("msgs", alice.ID(), charlie.ID()), nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusUnauthorized, code)
	require.Equal(t, `{"error":{"code":401,"message":"missing Authorization header"}}`, body)

	// GET /msgs/:kid/:rid
	req, err = api.NewRequest("GET", docs.Path("msgs", alice.ID(), charlie.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"msgs":[],"idx":0}`, body)

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
	req, err = api.NewRequest("GET", docs.Path("msgs", randKey.ID(), charlie.ID()), nil, "", clock.Now(), randKey)
	require.NoError(t, err)
	req.Header.Set("Authorization", randKey.ID().String()+":"+sig)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"verify failed"}}`, body)

	// POST /msgs/:kid/:rid (invalid recipient)
	content := []byte("test1")
	contentHash := encoding.EncodeBase64(keys.SHA256(content))
	req, err = api.NewRequest("POST", docs.Path("msgs", bob.ID(), charlie.ID()), bytes.NewReader(content), contentHash, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)
}
