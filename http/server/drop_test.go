package server_test

import (
	"bytes"
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestDrop(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	testDrop(t, env, testKeysSeeded())
}

func testDrop(t *testing.T, env *env, tk testKeys) {
	srv := newTestServer(t, env)
	clock := env.clock

	bob := tk.bob

	// PUT /drop/auth/:bob
	drop := url.Values{}
	drop.Set("token", "token1")
	req, err := http.NewAuthRequest("PUT", dstore.Path("/drop/auth", bob.ID()), strings.NewReader(drop.Encode()), http.ContentHash([]byte(drop.Encode())), clock.Now(), bob)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// POST /drop/:bob (with token)
	req, err = http.NewRequest("POST", dstore.Path("drop", bob.ID()), bytes.NewReader([]byte("hi")))
	req.Header.Set("Authorization", "token1")
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /drop/:kid (bob)
	req, err = http.NewAuthRequest("GET", dstore.Path("drop", bob.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var msgsResp api.MessagesResponse
	err = json.Unmarshal([]byte(body), &msgsResp)
	require.NoError(t, err)
	require.Equal(t, int64(1), msgsResp.Index)
	require.Equal(t, 1, len(msgsResp.Messages))
	require.Equal(t, []byte("hi"), msgsResp.Messages[0].Data)

	// POST /drop/:bob
	req, err = http.NewRequest("POST", dstore.Path("drop", bob.ID()), bytes.NewReader([]byte("content")))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)

	// POST /drop/:bob (with invalid token)
	req, err = http.NewRequest("POST", dstore.Path("drop", bob.ID()), bytes.NewReader([]byte("content")))
	req.Header.Set("Authorization", "invalidtoken1")
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
}

func TestDropFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	// firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	env := newEnvWithFire(t, testFirestore(t), tsutil.NewTestClock())
	// env.logLevel = server.DebugLevel
	testDrop(t, env, testKeysRandom())
}
