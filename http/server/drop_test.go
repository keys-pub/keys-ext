package server_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/saltpack"
	"github.com/stretchr/testify/require"
)

func TestDrops(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel

	tk := testKeysSeeded()
	alice, bob, channel := tk.alice, tk.bob, tk.channel

	srv := newTestServer(t, env)
	clock := env.clock

	// GET /drop/:kid
	req, err := http.NewAuthRequest("GET", dstore.Path("drop", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{"drops":[]}`, body)
	require.Equal(t, http.StatusOK, code)

	// Create token (alice)
	req, err = http.NewAuthRequest("POST", dstore.Path("user", alice.ID(), "token"), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var tokenResp api.UserTokenResponse
	err = json.Unmarshal([]byte(body), &tokenResp)
	require.NoError(t, err)
	token := tokenResp.Token
	require.NotEmpty(t, token)

	// POST /drop/:kid (alice share token with bob)
	tokenDrop := api.NewTokenDrop(token, alice.ID())
	drop, err := api.Encrypt(tokenDrop, alice, bob.ID())
	require.NoError(t, err)
	req, err = http.NewRequest("POST", dstore.Path("drop", bob.ID()), bytes.NewReader(drop))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// POST /drop/:kid (bob invites alice to channel)
	channelDrop := api.NewChannelDrop(channel, bob.ID())
	drop, err = api.Encrypt(channelDrop, bob, alice.ID())
	require.NoError(t, err)
	req, err = http.NewRequest("POST", dstore.Path("drop", alice.ID()), bytes.NewReader(drop))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /drop/:kid
	req, err = http.NewAuthRequest("GET", dstore.Path("drop", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.DropsResponse
	err = json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Drops))
	tokenDropOut, err := api.DecryptDrop(resp.Drops[0].Data, saltpack.NewKeyring(alice))
	require.NoError(t, err)
	require.Equal(t, tokenDrop, tokenDropOut)

	// DELETE /drop/:kid
	req, err = http.NewAuthRequest("DELETE", dstore.Path("drop", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /drop/:kid
	req, err = http.NewAuthRequest("GET", dstore.Path("drop", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"drops":[]}`, body)
	require.Equal(t, http.StatusOK, code)
}
