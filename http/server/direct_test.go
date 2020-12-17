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

func TestDirect(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	testDirect(t, env, testKeysSeeded())
}

func TestDirectFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	// firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	env := newEnvWithFire(t, testFirestore(t), tsutil.NewTestClock())
	// env.logLevel = server.DebugLevel
	testDirect(t, env, testKeysRandom())
}

func testDirect(t *testing.T, env *env, tk testKeys) {
	srv := newTestServer(t, env)
	clock := env.clock

	alice, bob := tk.alice, tk.bob

	// POST /dm/:bob/:alice (alice to bob)
	content := []byte("test1")
	contentHash := http.ContentHash(content)
	req, err := http.NewAuthRequest("POST", dstore.Path("dm", bob.ID(), alice.ID()), bytes.NewReader(content), contentHash, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)

	// POST /follow/:bob/:alice (bob follow alice)
	req, err = http.NewAuthRequest("POST", dstore.Path("follow", bob.ID(), alice.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// POST /dm/:bob/:alice (alice to bob, trying again after follow)
	req, err = http.NewAuthRequest("POST", dstore.Path("dm", bob.ID(), alice.ID()), bytes.NewReader(content), contentHash, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /dm/:kid (bob)
	req, err = http.NewAuthRequest("GET", dstore.Path("dm", bob.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var msgsResp api.MessagesResponse
	err = json.Unmarshal([]byte(body), &msgsResp)
	require.NoError(t, err)
	require.Equal(t, int64(1), msgsResp.Index)
	require.Equal(t, 1, len(msgsResp.Messages))
	require.Equal(t, []byte("test1"), msgsResp.Messages[0].Data)

	// DELETE /follow/:bob/:alice (bob unfollow alice)
	req, err = http.NewAuthRequest("DELETE", dstore.Path("follow", bob.ID(), alice.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// POST /dm/:bob/:alice (alice to bob, after unfollow)
	req, err = http.NewAuthRequest("POST", dstore.Path("dm", bob.ID(), alice.ID()), bytes.NewReader(content), contentHash, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
}
