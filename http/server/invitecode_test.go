package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/stretchr/testify/require"
)

func TestInviteCode(t *testing.T) {
	// api.SetLogger(api.NewLogger(api.DebugLevel))

	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	// POST /invite/:kid/:rid (alice, charlie)
	req, err := http.NewAuthRequest("POST", dstore.Path("/invite/code", alice.ID(), charlie.ID()), nil, "", env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var created api.InviteCodeCreateResponse
	err = json.Unmarshal([]byte(string(body)), &created)
	require.NoError(t, err)
	require.NotEmpty(t, created.Code)

	inviteCode := created.Code

	// GET /invite?code=..
	req, err = http.NewAuthRequest("GET", fmt.Sprintf("/invite/code/%s", url.QueryEscape(inviteCode)), nil, "", env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected := `{"sender":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","recipient":"kex1a4yj333g68pvd6hfqvufqkv4vy54jfe6t33ljd3kc9rpfty8xlgs2u3qxr"}`
	require.Equal(t, expected, string(body))
	var invite api.InviteCodeResponse
	err = json.Unmarshal([]byte(string(body)), &invite)
	require.NoError(t, err)
	require.Equal(t, charlie.ID(), invite.Recipient)

	// GET /invite?code=.. (bob, invalid)
	req, err = http.NewAuthRequest("GET", fmt.Sprintf("/invite/code/%s", url.QueryEscape(inviteCode)), nil, "", env.clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"code not found"}}`, string(body))
}
