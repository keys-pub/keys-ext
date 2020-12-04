package server_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	testMessages(t, env, testKeysSeeded())
}

func TestMessagesFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	// firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	env := newEnvWithFire(t, testFirestore(t), tsutil.NewTestClock())
	// env.logLevel = server.DebugLevel
	testMessages(t, env, testKeysRandom())
}

func testMessages(t *testing.T, env *env, tk testKeys) {
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))

	srv := newTestServer(t, env)
	clock := env.clock

	alice, channel := tk.alice, tk.channel
	aliceChannel := http.AuthKeys(
		http.NewAuthKey("Authorization", alice),
		http.NewAuthKey("Authorization-Channel", channel))

	// GET /channel/:cid/msgs (not found)
	req, err := http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
	require.Equal(t, http.StatusForbidden, code)

	// POST /channel/:cid/msgs (not found)
	content := []byte("test1")
	contentHash := http.ContentHash(content)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "msgs"), bytes.NewReader(content), contentHash, clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
	require.Equal(t, http.StatusForbidden, code)

	// PUT /channel/:cid
	req, err = http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/msgs
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"msgs":[],"idx":0}`, body)
	require.Equal(t, http.StatusOK, code)

	// POST /channel/:cid/msgs (no body)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "msgs"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected := `{"error":{"code":400,"message":"missing body"}}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusBadRequest, code)

	// POST /channel/:cid/msgs
	content = []byte("test1")
	contentHash = http.ContentHash(content)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "msgs"), bytes.NewReader(content), contentHash, clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)
	content2 := []byte("test2")
	contentHash2 := http.ContentHash(content2)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "msgs"), bytes.NewReader(content2), contentHash2, clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// PUT /channel/:cid/msgs (invalid method)
	req, err = http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID(), "msgs"), bytes.NewReader(content), contentHash, clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusMethodNotAllowed, code)
	require.Equal(t, `{"error":{"code":405,"message":"method not allowed"}}`, body)

	// GET /channel/:cid/msgs
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs")+"?limit=1", nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Index)
	require.Equal(t, 1, len(resp.Messages))
	require.Equal(t, []byte("test1"), resp.Messages[0].Data)

	// GET /channel/:cid/msgs?idx=next
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs")+"?idx="+strconv.Itoa(int(resp.Index)), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp2 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp2)
	require.NoError(t, err)
	require.Equal(t, int64(2), resp2.Index)
	require.Equal(t, 1, len(resp2.Messages))
	require.Equal(t, []byte("test2"), resp2.Messages[0].Data)

	// POST /channel/:cid/msgs
	content3 := []byte("test3")
	contentHash3 := http.ContentHash(content3)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "msgs"), bytes.NewReader(content3), contentHash3, clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/msgs
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs"), nil, "", clock.Now(), aliceChannel)
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

	// GET /channel/:cid/msgs (descending, limit=2)
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs")+"?dir=desc&limit=2", nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp4 api.MessagesResponse
	err = json.Unmarshal([]byte(body), &resp4)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp4.Messages))
	require.Equal(t, []byte("test3"), resp4.Messages[0].Data)
	require.Equal(t, []byte("test2"), resp4.Messages[1].Data)

	// POST /channel/:cid/msgs (message too large)
	large := bytes.Repeat([]byte{0x01}, 65*1024)
	largeHash := http.ContentHash(large)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "msgs"), bytes.NewReader(large), largeHash, clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusRequestEntityTooLarge, code)
	require.Equal(t, `{"error":{"code":413,"message":"request too large"}}`, body)
}

func TestMessagesAuth(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	env := newEnv(t)
	srv := newTestServer(t, env)
	clock := env.clock

	tk := testKeysSeeded()
	alice, channel := tk.alice, tk.channel
	aliceChannel := http.AuthKeys(
		http.NewAuthKey("Authorization", alice),
		http.NewAuthKey("Authorization-Channel", channel))

	// PUT /channel/:cid
	req, err := http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/msgs
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"msgs":[],"idx":0}`, body)

	// GET /channel/:cid/msgs (no auth)
	req, err = http.NewRequest("GET", dstore.Path("channel", channel.ID(), "msgs"), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)

	// GET /channel/:cid/msgs
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs"), nil, "", clock.Now(), aliceChannel)
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
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)

	// GET /channel/:cid/msgs (invalid authorization)
	authHeader := req.Header.Get("Authorization")
	randKey := keys.GenerateEdX25519Key()
	sig := strings.Split(authHeader, ":")[1]
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs"), nil, "", clock.Now(), http.Authorization(randKey))
	require.NoError(t, err)
	req.Header.Set("Authorization", randKey.ID().String()+":"+sig)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)

	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs"), nil, "", clock.Now(), http.Authorization(randKey))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)

	// POST /channel/:cid/msgs (invalid authorization)
	msg := []byte("test")
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "msgs"), bytes.NewReader(msg), http.ContentHash(msg), clock.Now(), http.Authorization(randKey))
	require.NoError(t, err)
	req.Header.Set("Authorization", randKey.ID().String()+":"+sig)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)

	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "msgs"), nil, "", clock.Now(), http.Authorization(randKey))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
}
