package server

import (
	"bytes"
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
	uc := keys.NewTestUserContext(rq, clock.Now)
	srv := newTestServer(t, clock, fi, uc)

	// GET /search
	req, err := http.NewRequest("GET", "/search", nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[]}`, body)

	// Alice
	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)

	// Alice sign user statement
	st := userMock(t, uc, alice, "alice", "github", clock, rq)
	// PUT alice
	req, err = http.NewRequest("PUT", st.URLPath(), bytes.NewReader(st.Bytes()))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "", body)

	// Bob
	bob, err := keys.NewKeyFromSeedPhrase(bobSeed, false)
	require.NoError(t, err)
	bobSc := keys.GenerateSigchain(bob, clock.Now())
	bobSt := bobSc.Statements()[0]
	// PUT bob
	req, err = http.NewRequest("PUT", bobSt.URLPath(), bytes.NewReader(bobSt.Bytes()))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "", body)

	// GET /search
	req, err = http.NewRequest("GET", "/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","users":[{"status":"ok","ts":"2009-02-13T15:31:30.005-08:00","user":{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","name":"alice","seq":1,"service":"github","url":"https://gist.github.com/alice/1"},"vts":"2009-02-13T15:31:30.006-08:00"}]},{"kid":"KNLPD1zD35FpXxP8q2B7JEWVqeJTxYH5RQKtGgrgNAtU"}]}`, body)

	// GET /search?q=alice
	req, err = http.NewRequest("GET", "/search?q=alice", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","users":[{"status":"ok","ts":"2009-02-13T15:31:30.005-08:00","user":{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","name":"alice","seq":1,"service":"github","url":"https://gist.github.com/alice/1"},"vts":"2009-02-13T15:31:30.006-08:00"}]}]}`, body)

	// GET /search?q=alice@github
	req, err = http.NewRequest("GET", "/search?q=alice@github", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","users":[{"status":"ok","ts":"2009-02-13T15:31:30.005-08:00","user":{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","name":"alice","seq":1,"service":"github","url":"https://gist.github.com/alice/1"},"vts":"2009-02-13T15:31:30.006-08:00"}]}]}`, body)

	// GET /search?q=unknown
	req, err = http.NewRequest("GET", "/search?q=unknown", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[]}`, body)

	// GET /search?q=KNLPD1zD35F (bob)
	req, err = http.NewRequest("GET", "/search?q=KNLPD1zD35F", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"KNLPD1zD35FpXxP8q2B7JEWVqeJTxYH5RQKtGgrgNAtU"}]}`, body)

	// GET /search?q=HX7DWqV9FtkXWJ (alice)
	req, err = http.NewRequest("GET", "/search?q=HX7DWqV9FtkXWJ", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","users":[{"status":"ok","ts":"2009-02-13T15:31:30.005-08:00","user":{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","name":"alice","seq":1,"service":"github","url":"https://gist.github.com/alice/1"},"vts":"2009-02-13T15:31:30.006-08:00"}]}]}`, body)
}
