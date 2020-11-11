package server_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/stretchr/testify/require"
)

func TestChannel(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel

	tk := testKeysSeeded()
	alice, _, channel, frank := tk.alice, tk.bob, tk.channel, tk.frank

	aliceChannel := http.AuthKeys(
		http.NewAuthKey("Authorization", alice),
		http.NewAuthKey("Authorization-Channel", channel))

	frankChannel := http.AuthKeys(
		http.NewAuthKey("Authorization", frank),
		http.NewAuthKey("Authorization-Channel", channel))

	srv := newTestServer(t, env)
	clock := env.clock

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

	// PUT /channel/:cid/info
	content := []byte("encryptedchannelinfo")
	contentHash := http.ContentHash(content)
	req, err = http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID(), "info"), bytes.NewReader(content), contentHash, clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/info (alice)
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "info"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, "encryptedchannelinfo", body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/info (frank)
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "info"), nil, "", clock.Now(), frankChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
	require.Equal(t, http.StatusForbidden, code)

	// // POST /channel/:cid/members
	// addMember := api.ChannelMembersAddRequest{
	// 	Members: []*api.ChannelMember{&api.ChannelMember{ID: bob.ID()}},
	// }
	// content, err := json.Marshal(addMember)
	// require.NoError(t, err)
	// contentHash := http.ContentHash(content)
	// req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "members"), bytes.NewReader(content), contentHash, clock.Now(), aliceChannel)
	// require.NoError(t, err)
	// code, _, body = srv.Serve(req)
	// require.Equal(t, `{}`, body)
	// require.Equal(t, http.StatusOK, code)

	// // GET /channel/:cid/members
	// req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "members"), nil, "", clock.Now(), aliceChannel)
	// require.NoError(t, err)
	// code, _, body = srv.Serve(req)
	// expected := `{"members":[{"kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","from":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077"},{"kid":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg","from":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077"}]}` + "\n"
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
	inviteBob := &api.ChannelInvite{
		CID:          channel.ID(),
		Recipient:    bob.ID(),
		Sender:       alice.ID(),
		EncryptedKey: []byte("testkey"),
	}
	content, err := json.Marshal(inviteBob)
	require.NoError(t, err)
	contentHash := http.ContentHash(content)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "invite"), bytes.NewReader(content), contentHash, clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// POST /channel/:cid/invite (alice invite frank)
	inviteFrank := &api.ChannelInvite{
		CID:          channel.ID(),
		Recipient:    frank.ID(),
		Sender:       alice.ID(),
		EncryptedKey: []byte("testkey"),
	}
	content, err = json.Marshal(inviteFrank)
	require.NoError(t, err)
	contentHash = http.ContentHash(content)
	req, err = http.NewAuthRequest("POST", dstore.Path("channel", channel.ID(), "invite"), bytes.NewReader(content), contentHash, clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/invites
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "invites"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected := `{"invites":[{"cid":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","recipient":"kex132r4llc7kwz9z4m6e4d0aeq9g4jk3htu38sfpp36q4tmc7h5nutsv4zjrd","sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","k":"dGVzdGtleQ=="},{"cid":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","recipient":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg","sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","k":"dGVzdGtleQ=="}]}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// GET /inbox/:kid/invites (bob)
	req, err = http.NewAuthRequest("GET", dstore.Path("inbox", bob.ID(), "invites"), nil, "", clock.Now(), http.Authorization(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected = `{"invites":[{"cid":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","recipient":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg","sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","k":"dGVzdGtleQ=="}]}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// POST /inbox/:kid/invite/:cid/accept (bob accept)
	req, err = http.NewAuthRequest("POST", dstore.Path("inbox", bob.ID(), "invite", channel.ID(), "accept"), nil, "", clock.Now(), bobChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /channel/:cid/members
	req, err = http.NewAuthRequest("GET", dstore.Path("channel", channel.ID(), "members"), nil, "", clock.Now(), aliceChannel)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected = `{"members":[{"kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","cid":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","from":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077"},{"kid":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg","cid":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","from":"kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg"}]}` + "\n"
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// GET /inbox/:kid/invites (frank)
	req, err = http.NewAuthRequest("GET", dstore.Path("inbox", frank.ID(), "invites"), nil, "", clock.Now(), http.Authorization(frank))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected = `{"invites":[{"cid":"kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep","recipient":"kex132r4llc7kwz9z4m6e4d0aeq9g4jk3htu38sfpp36q4tmc7h5nutsv4zjrd","sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","k":"dGVzdGtleQ=="}]}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// DELETE /inbox/:kid/invites/:cid (frank delete)
	req, err = http.NewAuthRequest("DELETE", dstore.Path("inbox", frank.ID(), "invite", channel.ID()), nil, "", clock.Now(), http.Authorization(frank))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /inbox/:kid/invites (frank)
	req, err = http.NewAuthRequest("GET", dstore.Path("inbox", frank.ID(), "invites"), nil, "", clock.Now(), http.Authorization(frank))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	expected = `{"invites":[]}`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)

	// GET /inbox/:kid/invites (bad auth)
	req, err = http.NewAuthRequest("GET", dstore.Path("inbox", alice.ID(), "invites"), nil, "", clock.Now(), http.Authorization(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
	require.Equal(t, http.StatusForbidden, code)
}
