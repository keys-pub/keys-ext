package server_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
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
	var create api.ChannelCreateResponse
	err = json.Unmarshal([]byte(body), &create)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.NotEmpty(t, create.Channel.Token)
	require.Equal(t, channel.ID(), create.Channel.ID)

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
	var channelOut api.Channel
	testJSONUnmarshal(t, []byte(body), &channelOut)
	require.Equal(t, keys.ID("kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep"), channelOut.ID)
	require.Equal(t, int64(1234567890004), channelOut.Timestamp)
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

	// POST /channels/status
	statusReq := api.ChannelsStatusRequest{
		Channels: map[keys.ID]string{channel.ID(): create.Channel.Token},
	}
	req, err = http.NewRequest("POST", "/channels/status", bytes.NewReader(testJSONMarshal(t, statusReq)))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	var statusResp api.ChannelsStatusResponse
	testJSONUnmarshal(t, []byte(body), &statusResp)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, keys.ID("kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep"), statusResp.Channels[0].ID)
	require.Equal(t, int64(1234567890004), statusResp.Channels[0].Timestamp)
	require.Equal(t, int64(1), statusResp.Channels[0].Index)
}
