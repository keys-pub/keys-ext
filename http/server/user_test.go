package server

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestUserSearch(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	users := testUserStore(t, fi, rq, clock)
	srv := newTestServer(t, clock, fi, users)

	// GET /user/search
	req, err := http.NewRequest("GET", "/user/search", nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[]}`, body)

	// Alice
	alice := keys.NewEd25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// Alice sign user statement
	st := userMock(t, users, alice, "alice", "github", rq)
	// PUT alice
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(st.Bytes()))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "", body)

	// GET /user/search
	req, err = http.NewRequest("GET", "/user/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"kse132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawquwc7vw","users":{"status":"ok","ts":1234567890005,"user":{"k":"kse132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawquwc7vw","n":"alice","sq":1,"sr":"github","u":"https://gist.github.com/alice/1"},"vts":1234567890006}}]}`, body)

	// GET /user/search?q=alice
	req, err = http.NewRequest("GET", "/user/search?q=alice", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"kse132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawquwc7vw","users":{"status":"ok","ts":1234567890005,"user":{"k":"kse132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawquwc7vw","n":"alice","sq":1,"sr":"github","u":"https://gist.github.com/alice/1"},"vts":1234567890006}}]}`, body)

	// GET /user/search?q=alice@github
	req, err = http.NewRequest("GET", "/user/search?q=alice@github", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"kse132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawquwc7vw","users":{"status":"ok","ts":1234567890005,"user":{"k":"kse132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawquwc7vw","n":"alice","sq":1,"sr":"github","u":"https://gist.github.com/alice/1"},"vts":1234567890006}}]}`, body)

	// GET /user/search?q=unknown
	req, err = http.NewRequest("GET", "/user/search?q=unknown", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[]}`, body)

	// GET /user/kse132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawquwc7vw (alice)
	req, err = http.NewRequest("GET", "/user/kse132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawquwc7vw", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"result":{"status":"ok","ts":1234567890005,"user":{"k":"kse132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawquwc7vw","n":"alice","sq":1,"sr":"github","u":"https://gist.github.com/alice/1"},"vts":1234567890006}}`, body)
}
