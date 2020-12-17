package server_test

import (
	"bytes"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/stretchr/testify/require"
)

func TestChannel(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	testChannel(t, env, testKeysSeeded())
}

// func TestChannelFirestore(t *testing.T) {
// 	if os.Getenv("TEST_FIRESTORE") != "1" {
// 		t.Skip()
// 	}
// 	// firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
// 	env := newEnvWithFire(t, testFirestore(t), tsutil.NewTestClock())
// 	// env.logLevel = server.DebugLevel
// 	testChannel(t, env, testKeysRandom())
// }

func testChannel(t *testing.T, env *env, tk testKeys) {
	channel := tk.channel

	srv := newTestServer(t, env)
	clock := env.clock
	randKey := keys.GenerateEdX25519Key()

	// PUT /channel/:cid
	req, err := http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), channel)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// PUT /channel/:cid (already exists)
	req, err = http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), channel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":409,"message":"channel already exists"}}`, body)
	require.Equal(t, http.StatusConflict, code)

	// GET /channel/:cid
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), channel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"id":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","ts":1234567890004}`+"\n", body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid (not found, forbidden)
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", randKey.ID()), nil, "", clock.Now(), channel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
	require.Equal(t, http.StatusForbidden, code)

	// POST /channel/:cid/msgs
	msg := []byte("test1")
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "msgs"), bytes.NewReader(msg), http.ContentHash(msg), clock.Now(), channel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /channel/:cid
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), channel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"id":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","idx":1,"ts":1234567890004}`+"\n", body)
	require.Equal(t, http.StatusOK, code)
}
