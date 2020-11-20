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
	alice, _, channel := tk.alice, tk.bob, tk.channel

	aliceChannel := http.AuthKeys(
		http.NewAuthKey("Authorization", alice),
		http.NewAuthKey("Authorization-Channel", channel))

	srv := newTestServer(t, env)
	clock := env.clock
	randKey := keys.GenerateEdX25519Key()

	// PUT /channel/:cid
	req, err := http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// PUT /channel/:cid (already exists)
	req, err = http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":409,"message":"channel already exists"}}`, body)
	require.Equal(t, http.StatusConflict, code)

	// GET /channel/:cid
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"id":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","creator":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","ts":1234567890005}`+"\n", body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid (not found, forbidden)
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", randKey.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
	require.Equal(t, http.StatusForbidden, code)

	// POST /channel/:cid/msgs
	content := []byte("test1")
	contentHash := http.ContentHash(content)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "msgs"), bytes.NewReader(content), contentHash, clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /channel/:cid
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"id":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","creator":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","idx":1,"ts":1234567890005}`+"\n", body)
	require.Equal(t, http.StatusOK, code)

	// // POST /channel/:cid/users
	// addUser := api.ChannelUsersAddRequest{
	// 	Users: []*api.ChannelUser{&api.ChannelUser{ID: bob.ID()}},
	// }
	// content, err := json.Marshal(addUser)
	// require.NoError(t, err)
	// contentHash := http.ContentHash(content)
	// req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "users"), bytes.NewReader(content), contentHash, clock.Now(), aliceChannel)
	// require.NoError(t, err)
	// code, _, body = srv.Serve(req)
	// require.Equal(t, `{}`, body)
	// require.Equal(t, http.StatusOK, code)

	// // GET /channel/:cid/users
	// req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "users"), nil, "", clock.Now(), aliceChannel)
	// require.NoError(t, err)
	// code, _, body = srv.Serve(req)
	// expected := `{"users":[{"kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","from":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077"},{"kid":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg","from":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077"}]}` + "\n"
	// require.Equal(t, expected, body)
	// require.Equal(t, http.StatusOK, code)
}

func TestChannelInvite(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel

	tk := testKeysSeeded()
	alice, bob, channel, frank := tk.alice, tk.bob, tk.channel, tk.frank

	aliceChannel := http.AuthKeys(
		http.NewAuthKey("Authorization", alice),
		http.NewAuthKey("Authorization-Channel", channel))

	bobChannel := http.AuthKeys(
		http.NewAuthKey("Authorization", bob),
		http.NewAuthKey("Authorization-Channel", channel))

	srv := newTestServer(t, env)
	clock := env.clock

	// PUT /channel/:cid
	req, err := http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// POST /channel/:cid/invite (alice invite bob)
	invitesBob := []*api.ChannelInvite{
		&api.ChannelInvite{
			Channel:      channel.ID(),
			Recipient:    bob.ID(),
			Sender:       alice.ID(),
			EncryptedKey: []byte("testkey"),
		},
	}
	invites, err := json.Marshal(invitesBob)
	require.NoError(t, err)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "invites"), bytes.NewReader(invites), http.ContentHash(invites), clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// POST /channel/:cid/invite (alice invite frank)
	invitesFrank := []*api.ChannelInvite{
		&api.ChannelInvite{
			Channel:      channel.ID(),
			Recipient:    frank.ID(),
			Sender:       alice.ID(),
			EncryptedKey: []byte("testkey"),
		},
	}
	invites, err = json.Marshal(invitesFrank)
	require.NoError(t, err)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "invites"), bytes.NewReader(invites), http.ContentHash(invites), clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/invites
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "invites"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected := `{"invites":[{"channel":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","recipient":"kex132r4llc7kwz9z4m6e4d0aeq9g4jk3htu38sfpp36q4tmc7h5nutsv4zjrd","sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","k":"dGVzdGtleQ=="},{"channel":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","recipient":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg","sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","k":"dGVzdGtleQ=="}]}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// GET /user/:kid/invites (bob)
	req, err = http.NewAuthRequest("GET", dstore.Path("user", bob.ID(), "invites"), nil, "", clock.Now(), http.Authorization(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected = `{"invites":[{"channel":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","recipient":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg","sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","k":"dGVzdGtleQ=="}]}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// GET /user/:kid/invite/:cid (bob gets invite)
	req, err = http.NewAuthRequest("GET", dstore.Path("user", bob.ID(), "invite", channel.ID()), nil, "", clock.Now(), http.Authorization(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected = `{"invite":{"channel":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","recipient":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg","sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","k":"dGVzdGtleQ=="}}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// POST /user/:kid/invite/:cid/accept (bob accept)
	req, err = http.NewAuthRequest("POST", dstore.Path("user", bob.ID(), "invite", channel.ID(), "accept"), nil, "", clock.Now(), bobChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/users
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "users"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected = `{"users":[{"channel":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","user":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","from":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077"},{"channel":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","user":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg","from":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg"}]}` + "\n"
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// GET /user/:kid/invites (frank)
	req, err = http.NewAuthRequest("GET", dstore.Path("user", frank.ID(), "invites"), nil, "", clock.Now(), http.Authorization(frank))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected = `{"invites":[{"channel":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","recipient":"kex132r4llc7kwz9z4m6e4d0aeq9g4jk3htu38sfpp36q4tmc7h5nutsv4zjrd","sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","k":"dGVzdGtleQ=="}]}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// DELETE /user/:kid/invites/:cid (frank delete)
	req, err = http.NewAuthRequest("DELETE", dstore.Path("user", frank.ID(), "invite", channel.ID()), nil, "", clock.Now(), http.Authorization(frank))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /user/:kid/invites (frank)
	req, err = http.NewAuthRequest("GET", dstore.Path("user", frank.ID(), "invites"), nil, "", clock.Now(), http.Authorization(frank))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected = `{"invites":[]}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// GET /user/:kid/invites (bad auth)
	req, err = http.NewAuthRequest("GET", dstore.Path("user", alice.ID(), "invites"), nil, "", clock.Now(), http.Authorization(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
	require.Equal(t, http.StatusForbidden, code)
}
