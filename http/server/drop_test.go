package server_test

import (
	"bytes"
	"encoding/json"
	"os"
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

	alice, bob := tk.alice, tk.bob

	// PUT /follow/:bob/:alice
	req, err := http.NewAuthRequest("PUT", dstore.Path("follow", bob.ID(), alice.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// POST /drop/:alice/:bob
	req, err = http.NewAuthRequest("POST", dstore.Path("drop", alice.ID(), bob.ID()), bytes.NewReader([]byte("hi")), http.ContentHash([]byte("hi")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /drop/:bob
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

	// POST /drop/:bob/:alice (alice doesn't follow bob)
	req, err = http.NewAuthRequest("POST", dstore.Path("drop", bob.ID(), alice.ID()), bytes.NewReader([]byte("hi")), http.ContentHash([]byte("hi")), clock.Now(), bob)
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
