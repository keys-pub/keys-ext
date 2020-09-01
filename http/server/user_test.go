package server_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/user"
	"github.com/stretchr/testify/require"
)

func TestUserSearch(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)

	// GET /users/search
	req, err := http.NewRequest("GET", "/users/search", nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[]}`, body)

	// Alice
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// Alice statement
	sc := keys.NewSigchain(alice.ID())
	st, err := user.MockStatement(alice, sc, "alice", "github", env.req, env.clock)
	require.NoError(t, err)

	// PUT alice
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /users/search
	req, err = http.NewRequest("GET", "/users/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890004,"ts":1234567890004}]}`, body)

	// GET /users/search?q=alice
	req, err = http.NewRequest("GET", "/users/search?q=alice", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890004,"ts":1234567890004}]}`, body)

	// GET /users/search?q=kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077
	req, err = http.NewRequest("GET", "/users/search?q=kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890004,"ts":1234567890004,"mf":"kid"}]}`, body)

	// GET /users/search?q=kbx1rvd43h2sag2tvrdp0duse5p82nvhpjd6hpjwhv7q7vqklega8atshec5ws
	req, err = http.NewRequest("GET", "/users/search?q=kbx1rvd43h2sag2tvrdp0duse5p82nvhpjd6hpjwhv7q7vqklega8atshec5ws", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890004,"ts":1234567890004,"mf":"kid"}]}`, body)

	// GET /users/search?q=alice@github
	req, err = http.NewRequest("GET", "/users/search?q=alice@github", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890004,"ts":1234567890004}]}`, body)

	// GET /users/search?q=unknown
	req, err = http.NewRequest("GET", "/users/search?q=unknown", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[]}`, body)
}

func TestUserGet(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)

	// Alice
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// Alice statement
	sc := keys.NewSigchain(alice.ID())
	st, err := user.MockStatement(alice, sc, "alice", "github", env.req, env.clock)
	require.NoError(t, err)

	// PUT alice
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /users/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 (alice)
	req, err = http.NewRequest("GET", "/users/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890004,"ts":1234567890004}]}`, body)

	// GET /users/:kid (not found)
	key := keys.GenerateEdX25519Key()
	req, err = http.NewRequest("GET", "/users/"+key.ID().String(), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"no users found"}}`, body)

	// GET /users/:kid (invalid)
	req, err = http.NewRequest("GET", "/users/testkey", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"invalid kid"}}`, body)

	// GET /users/kbx1rvd43h2sag2tvrdp0duse5p82nvhpjd6hpjwhv7q7vqklega8atshec5ws
	req, err = http.NewRequest("GET", "/users/kbx1rvd43h2sag2tvrdp0duse5p82nvhpjd6hpjwhv7q7vqklega8atshec5ws", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890004,"ts":1234567890004}]}`, body)

	// GET /user/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 (deprecated)
	req, err = http.NewRequest("GET", "/user/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"user":{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890004,"ts":1234567890004}}`, body)
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
	sc := keys.NewSigchain(alice.ID())
	st, err := user.MockStatement(alice, sc, "alice", "github", env.req, env.clock)
	require.NoError(t, err)
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /users/search
	req, err = http.NewRequest("GET", "/users/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890004,"ts":1234567890004}]}`, body)

	// PUT /sigchain/alice2/1
	sc2 := keys.NewSigchain(alice2.ID())
	st2, err := user.MockStatement(alice2, sc2, "alice", "github", env.req, env.clock)
	require.NoError(t, err)
	b2, err := st2.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice2.ID()), bytes.NewReader(b2))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusConflict, code)
	require.Equal(t, `{"error":{"code":409,"message":"user already exists with key kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077, if you removed or revoked the previous statement you may need to wait briefly for search to update"}}`, body)
}

func TestUserMultiple(t *testing.T) {
	// user.SetLogger(user.NewLogger(user.DebugLevel))

	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)

	// Alice
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	sc := keys.NewSigchain(alice.ID())

	// PUT /sigchain/alice/1
	st, err := user.MockStatement(alice, sc, "alice", "github", env.req, env.clock)
	require.NoError(t, err)
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /users/search
	req, err = http.NewRequest("GET", "/users/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890004,"ts":1234567890004}]}`, body)

	// PUT /sigchain/alice/2
	st2, err := user.MockStatement(alice, sc, "alice", "twitter", env.req, env.clock)
	require.NoError(t, err)
	b2, err := st2.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/2", alice.ID()), bytes.NewReader(b2))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /users/search
	req, err = http.NewRequest("GET", "/users/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[{"id":"alice@github","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"service":"github","url":"https://gist.github.com/alice/1","status":"ok","verifiedAt":1234567890012,"ts":1234567890012},{"id":"alice@twitter","name":"alice","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":2,"service":"twitter","url":"https://mobile.twitter.com/alice/status/1","status":"ok","verifiedAt":1234567890013,"ts":1234567890013}]}`, body)
}
