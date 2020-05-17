package server_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestUserSearch(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)

	// GET /user/search
	req, err := http.NewRequest("GET", "/user/search", nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[]}`, body)

	// Alice
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// Alice sign user statement
	st := userMock(t, env.users, alice, "alice", "github", env.req)
	// PUT alice
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /user/search
	req, err = http.NewRequest("GET", "/user/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890003}]}`, body)

	// GET /user/search?q=alice
	req, err = http.NewRequest("GET", "/user/search?q=alice", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890003}]}`, body)

	// GET /user/search?q=alice@github
	req, err = http.NewRequest("GET", "/user/search?q=alice@github", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890003}]}`, body)

	// GET /user/search?q=unknown
	req, err = http.NewRequest("GET", "/user/search?q=unknown", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[]}`, body)

	// GET /user/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 (alice)
	req, err = http.NewRequest("GET", "/user/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"user":{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890003}}`, body)

	// GET /user/:kid (not found)
	key := keys.GenerateEdX25519Key()
	req, err = http.NewRequest("GET", "/user/"+key.ID().String(), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"user not found"}}`, body)
}

func TestUserDuplicate(t *testing.T) {
	// user.SetLogger(user.NewLogger(user.DebugLevel))

	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)

	// Alice
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	alice2 := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x08}, 32)))

	// PUT /sigchain/alice/1
	st := userMock(t, env.users, alice, "alice", "github", env.req)
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /user/search
	req, err = http.NewRequest("GET", "/user/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890003}]}`, body)

	// PUT /sigchain/alice2/1
	st2 := userMock(t, env.users, alice2, "alice", "github", env.req)
	b2, err := st2.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice2.ID()), bytes.NewReader(b2))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusConflict, code)
	require.Equal(t, `{"error":{"code":409,"message":"user already exists with key kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077, if you removed or revoked the previous statement you may need to wait briefly for search to update"}}`, body)
}
