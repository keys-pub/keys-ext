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

func TestDirectMessages(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	testDirectMessages(t, env, testKeysSeeded())
}

func testDirectMessages(t *testing.T, env *env, tk testKeys) {
	srv := newTestServer(t, env)
	clock := env.clock

	alice, bob := tk.alice, tk.bob

	// PUT /follow/:bob/:alice
	req, err := http.NewAuthRequest("PUT", dstore.Path("follow", bob.ID(), alice.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))

	// POST /direct/:alice/:bob
	req, err = http.NewAuthRequest("POST", dstore.Path("dm", alice.ID(), bob.ID()), bytes.NewReader([]byte("hi")), http.ContentHash([]byte("hi")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, string(body))
	require.Equal(t, http.StatusOK, code)

	// GET /direct/:bob
	req, err = http.NewAuthRequest("GET", dstore.Path("dm", bob.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var msgsResp api.Events
	err = json.Unmarshal([]byte(string(body)), &msgsResp)
	require.NoError(t, err)
	require.Equal(t, int64(1), msgsResp.Index)
	require.Equal(t, 1, len(msgsResp.Events))
	require.Equal(t, []byte("hi"), msgsResp.Events[0].Data)

	// POST /direct/:bob/:alice (alice doesn't follow bob)
	req, err = http.NewAuthRequest("POST", dstore.Path("dm", bob.ID(), alice.ID()), bytes.NewReader([]byte("hi")), http.ContentHash([]byte("hi")), clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, string(body))
}

func TestDirectMessagesFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	// firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	env := newEnvWithFire(t, testFirestore(t), tsutil.NewTestClock())
	// env.logLevel = server.DebugLevel
	testDirectMessages(t, env, testKeysRandom())
}
