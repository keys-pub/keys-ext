package server

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	users := testUserStore(t, fi, rq, clock)
	srv := newTestServer(t, clock, fi, users)

	// GET /search
	req, err := http.NewRequest("GET", "/search", nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[]}`, body)

	// Alice
	alice, err := keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	require.NoError(t, err)

	// Alice sign user statement
	st := userMock(t, users, alice, "alice", "github", rq)
	// PUT alice
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(st.Bytes()))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "", body)

	// Bob
	bob, err := keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	require.NoError(t, err)
	bobSc := keys.NewSigchain(bob.PublicKey())
	bobSt, err := keys.GenerateStatement(bobSc, []byte("bob"), bob, "", clock.Now())
	require.NoError(t, err)
	err = bobSc.Add(bobSt)
	require.NoError(t, err)
	// PUT bob
	t.Logf("bob: %s", bob.ID())
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", bob.ID()), bytes.NewReader(bobSt.Bytes()))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "", body)

	// GET /search
	req, err = http.NewRequest("GET", "/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen","users":[{"status":"ok","ts":1234567890005,"user":{"k":"kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen","n":"alice","sq":1,"sr":"github","u":"https://gist.github.com/alice/1"},"vts":1234567890006}]},{"kid":"kpe1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2qrt73l9"}]}`, body)

	// GET /search?q=alice
	req, err = http.NewRequest("GET", "/search?q=alice", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen","users":[{"status":"ok","ts":1234567890005,"user":{"k":"kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen","n":"alice","sq":1,"sr":"github","u":"https://gist.github.com/alice/1"},"vts":1234567890006}]}]}`, body)

	// GET /search?q=alice@github
	req, err = http.NewRequest("GET", "/search?q=alice@github", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen","users":[{"status":"ok","ts":1234567890005,"user":{"k":"kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen","n":"alice","sq":1,"sr":"github","u":"https://gist.github.com/alice/1"},"vts":1234567890006}]}]}`, body)

	// GET /search?q=unknown
	req, err = http.NewRequest("GET", "/search?q=unknown", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[]}`, body)

	// GET /search?q=kpe1syuhwr4 (bob)
	req, err = http.NewRequest("GET", "/search?q=kpe1syuhwr4", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"kpe1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2qrt73l9"}]}`, body)

	// GET /search?q=kpe132yw (alice)
	req, err = http.NewRequest("GET", "/search?q=kpe132yw", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen","users":[{"status":"ok","ts":1234567890005,"user":{"k":"kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen","n":"alice","sq":1,"sr":"github","u":"https://gist.github.com/alice/1"},"vts":1234567890006}]}]}`, body)
}
