package server_test

import (
	"testing"

	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/stretchr/testify/require"
)

func TestInbox(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel

	tk := testKeysSeeded()
	alice, channel, channel2 := tk.alice, tk.channel, tk.channel2

	aliceChannel := http.AuthKeys(
		http.NewAuthKey("Authorization", alice),
		http.NewAuthKey("Authorization-Channel", channel))

	aliceChannel2 := http.AuthKeys(
		http.NewAuthKey("Authorization", alice),
		http.NewAuthKey("Authorization-Channel", channel2))

	srv := newTestServer(t, env)
	clock := env.clock

	// GET /inbox/:kid/channels
	req, err := http.NewAuthRequest("GET", dstore.Path("inbox", alice.ID(), "channels"), nil, "", clock.Now(), http.Authorization(alice))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{"channels":[]}`, body)
	require.Equal(t, http.StatusOK, code)

	// PUT /channel/:cid
	req, err = http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /inbox/:kid/channels
	req, err = http.NewAuthRequest("GET", dstore.Path("inbox", alice.ID(), "channels"), nil, "", clock.Now(), http.Authorization(alice))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"channels":[{"id":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep"}]}`, body)
	require.Equal(t, http.StatusOK, code)

	// PUT /channel/:cid
	req, err = http.NewAuthRequest("PUT", dstore.Path("channel", channel2.ID()), nil, "", clock.Now(), aliceChannel2)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /inbox/:kid/channels
	req, err = http.NewAuthRequest("GET", dstore.Path("inbox", alice.ID(), "channels"), nil, "", clock.Now(), http.Authorization(alice))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"channels":[{"id":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep"},{"id":"kex1tan3x22v8nc6s98gmr9q3zwmy0ngywm4yja0zdylh37e752jj3dsur2s3g"}]}`, body)
	require.Equal(t, http.StatusOK, code)
}
