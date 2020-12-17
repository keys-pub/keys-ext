package server_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestFollow(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	testFollow(t, env, testKeysSeeded())
}

func TestFollowFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	// firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	env := newEnvWithFire(t, testFirestore(t), tsutil.NewTestClock())
	// env.logLevel = server.DebugLevel
	testFollow(t, env, testKeysRandom())
}

func testFollow(t *testing.T, env *env, tk testKeys) {
	srv := newTestServer(t, env)
	clock := env.clock

	alice, bob := tk.alice, tk.bob

	// POST /follow/:bob/:alice (bob follow alice)
	req, err := http.NewAuthRequest("POST", dstore.Path("follow", bob.ID(), alice.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /follows/:bob
	req, err = http.NewAuthRequest("GET", dstore.Path("follows", bob.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var followsResp api.FollowsResponse
	err = json.Unmarshal([]byte(body), &followsResp)
	require.NoError(t, err)
	require.Equal(t, 1, len(followsResp.Follows))
	require.Equal(t, bob.ID(), followsResp.Follows[0].KID)
	require.Equal(t, alice.ID(), followsResp.Follows[0].User)

	// DELETE /follow/:bob/:alice (bob unfollow alice)
	req, err = http.NewAuthRequest("DELETE", dstore.Path("follow", bob.ID(), alice.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /follows/:bob
	req, err = http.NewAuthRequest("GET", dstore.Path("follows", bob.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"follows":[]}`, body)
}
