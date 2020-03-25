package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/stretchr/testify/require"
)

func TestEphem(t *testing.T) {
	// api.SetLogger(api.NewLogger(api.DebugLevel))

	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	// PUT /ephem/:kid/:rid?code=1 (alice to charlie)
	req, err := api.NewRequest("PUT", keys.Path("ephem", alice.ID(), charlie.ID())+"?code=1", bytes.NewReader([]byte("hi")), env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var ephem api.EphemResponse
	err = json.Unmarshal([]byte(body), &ephem)
	require.NoError(t, err)
	inviteCode := ephem.Code
	require.NotEmpty(t, inviteCode)

	// GET /invite?code=..
	req, err = api.NewRequest("GET", fmt.Sprintf("/invite?code=%s", url.QueryEscape(inviteCode)), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected := `{"sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","recipient":"kex1a4yj333g68pvd6hfqvufqkv4vy54jfe6t33ljd3kc9rpfty8xlgs2u3qxr"}`
	require.Equal(t, expected, body)
	var invite api.InviteResponse
	err = json.Unmarshal([]byte(body), &invite)
	require.NoError(t, err)
	require.Equal(t, charlie.ID(), invite.Recipient)

	// GET /ephem/:kid/:rid (charlie from alice)
	req, err = api.NewRequest("GET", keys.Path("ephem", invite.Recipient, invite.Sender), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "hi", body)

	// GET again
	req, err = api.NewRequest("GET", keys.Path("ephem", invite.Recipient, invite.Sender), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /ephem/:kid/:rid (alice to charlie)
	req, err = api.NewRequest("PUT", keys.Path("ephem", alice.ID(), charlie.ID()), bytes.NewReader([]byte("hi")), env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// DEL (invalid auth)
	req, err = api.NewRequest("DELETE", keys.Path("ephem", alice.ID(), charlie.ID()), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// DEL /ephem/:kid/:rid
	req, err = api.NewRequest("DELETE", keys.Path("ephem", alice.ID(), charlie.ID()), nil, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET (after delete)
	req, err = api.NewRequest("GET", keys.Path("ephem", invite.Recipient, invite.Sender), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)
}
