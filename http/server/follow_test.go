package server_test

import (
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

func TestFollow(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	testFollow(t, env, testKeysSeeded())
}

func testFollow(t *testing.T, env *env, tk testKeys) {
	srv := newTestServer(t, env)
	clock := env.clock

	alice, bob := tk.alice, tk.bob

	// PUT /follow/:bob/:alice (bob follow alice)
	follow := url.Values{}
	follow.Set("token", "token1")
	req, err := http.NewAuthRequest("PUT", dstore.Path("follow", bob.ID(), alice.ID()), strings.NewReader(follow.Encode()), http.ContentHash([]byte(follow.Encode())), clock.Now(), bob)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /follows/:alice
	req, err = http.NewAuthRequest("GET", dstore.Path("follows", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var followsResp api.FollowsResponse
	err = json.Unmarshal([]byte(body), &followsResp)
	require.NoError(t, err)
	require.Equal(t, 1, len(followsResp.Follows))
	require.Equal(t, alice.ID(), followsResp.Follows[0].Recipient)
	require.Equal(t, bob.ID(), followsResp.Follows[0].Sender)
	require.Equal(t, "token1", followsResp.Follows[0].Token)

	// GET /follow/:bob/:alice
	req, err = http.NewAuthRequest("GET", dstore.Path("follow", bob.ID(), alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var followResp api.FollowResponse
	err = json.Unmarshal([]byte(body), &followResp)
	require.NoError(t, err)
	require.Equal(t, alice.ID(), followResp.Follow.Recipient)
	require.Equal(t, bob.ID(), followResp.Follow.Sender)
	require.Equal(t, "token1", followResp.Follow.Token)

	// GET /follow/:bob/:alice (invalid auth)
	req, err = http.NewAuthRequest("GET", dstore.Path("follow", bob.ID(), alice.ID()), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)

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

func TestFollowFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	// firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	env := newEnvWithFire(t, testFirestore(t), tsutil.NewTestClock())
	// env.logLevel = server.DebugLevel
	testFollow(t, env, testKeysRandom())
}
