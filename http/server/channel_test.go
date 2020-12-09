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
	require.Equal(t, `{"id":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","ts":1234567890005}`+"\n", body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid (not found, forbidden)
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", randKey.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
	require.Equal(t, http.StatusForbidden, code)

	// POST /channel/:cid/msgs
	msg := []byte("test1")
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "msgs"), bytes.NewReader(msg), http.ContentHash(msg), clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET /channel/:cid
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"id":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","idx":1,"ts":1234567890005}`+"\n", body)
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
	info := &api.ChannelInfo{Name: "test"}
	inviteBob, err := api.NewChannelInvite(channel, info, alice, bob.ID())
	require.NoError(t, err)
	invitesBob := api.ChannelInvitesRequest{Invites: []*api.ChannelInvite{inviteBob}}
	invitesBobBody, err := json.Marshal(invitesBob)
	require.NoError(t, err)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "invites"), bytes.NewReader(invitesBobBody), http.ContentHash(invitesBobBody), clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// POST /channel/:cid/invite (alice invite frank)
	inviteFrank, err := api.NewChannelInvite(channel, info, alice, frank.ID())
	require.NoError(t, err)
	invitesFrank := api.ChannelInvitesRequest{Invites: []*api.ChannelInvite{inviteFrank}}
	invitesFrankBody, err := json.Marshal(invitesFrank)
	require.NoError(t, err)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "invites"), bytes.NewReader(invitesFrankBody), http.ContentHash(invitesFrankBody), clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/invites
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "invites"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var invitesResp api.ChannelInvitesResponse
	err = json.Unmarshal([]byte(body), &invitesResp)
	require.NoError(t, err)
	require.Equal(t, []*api.ChannelInvite{inviteFrank, inviteBob}, invitesResp.Invites)

	// GET /user/:kid/invites (bob)
	req, err = http.NewAuthRequest("GET", dstore.Path("user", bob.ID(), "invites"), nil, "", clock.Now(), http.Authorization(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	err = json.Unmarshal([]byte(body), &invitesResp)
	require.NoError(t, err)
	require.Equal(t, []*api.ChannelInvite{inviteBob}, invitesResp.Invites)

	// GET /user/:kid/invite/:cid (bob gets invite)
	req, err = http.NewAuthRequest("GET", dstore.Path("user", bob.ID(), "invite", channel.ID()), nil, "", clock.Now(), http.Authorization(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var userInviteResp api.UserChannelInviteResponse
	err = json.Unmarshal([]byte(body), &userInviteResp)
	require.NoError(t, err)
	require.Equal(t, inviteBob, userInviteResp.Invite)

	// PUT /user/:kid/channel/:cid (bob join)
	req, err = http.NewAuthRequest("PUT", dstore.Path("user", bob.ID(), "channel", channel.ID()), nil, "", clock.Now(), bobChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/users
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "users"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var usersResp api.ChannelUsersResponse
	err = json.Unmarshal([]byte(body), &usersResp)
	require.NoError(t, err)
	require.Equal(t, []*api.ChannelUser{
		&api.ChannelUser{
			Channel: channel.ID(),
			User:    alice.ID(),
		},
		&api.ChannelUser{
			Channel: channel.ID(),
			User:    bob.ID(),
		},
	}, usersResp.Users)

	// GET /user/:kid/invites (frank)
	req, err = http.NewAuthRequest("GET", dstore.Path("user", frank.ID(), "invites"), nil, "", clock.Now(), http.Authorization(frank))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	err = json.Unmarshal([]byte(body), &invitesResp)
	require.NoError(t, err)
	require.Equal(t, []*api.ChannelInvite{inviteFrank}, invitesResp.Invites)

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
	expected := `{"invites":[]}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/invites (alice)
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "invites"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	err = json.Unmarshal([]byte(body), &invitesResp)
	require.NoError(t, err)
	require.Equal(t, []*api.ChannelInvite{inviteBob}, invitesResp.Invites)

	// GET /user/:kid/invites (bad auth)
	req, err = http.NewAuthRequest("GET", dstore.Path("user", alice.ID(), "invites"), nil, "", clock.Now(), http.Authorization(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
	require.Equal(t, http.StatusForbidden, code)

	// DELETE /channel/:cid/invite/:kid (alice uininvite bob)
	req, err = http.NewAuthRequest("DELETE", dstore.Path("channel", channel.ID(), "invite", bob.ID()), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/invites (alice)
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "invites"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected = `{"invites":[]}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)
}
