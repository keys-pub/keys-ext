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
	srv := newTestServer(t, clock, fi)

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
	aliceSc := keys.NewSigchain(alice.PublicKey().SignPublicKey())
	usr, err := keys.NewUser(alice.ID(), "test", "alice", "test://", 1)
	require.NoError(t, err)
	aliceSt, err := keys.GenerateUserStatement(aliceSc, usr, alice.SignKey(), clock.Now())
	require.NoError(t, err)
	// PUT alice
	req, err = http.NewRequest("PUT", aliceSt.URLPath(), bytes.NewReader(aliceSt.Bytes()))
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
	require.Equal(t, `{"results":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","users":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","name":"alice","seq":1,"service":"test","url":"test:","ucts":"2009-02-13T15:31:30.005-08:00"}]},{"kid":"KNLPD1zD35FpXxP8q2B7JEWVqeJTxYH5RQKtGgrgNAtU"}]}`, body)

	// GET /search?q=alice
	req, err = http.NewRequest("GET", "/search?q=alice", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","users":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","name":"alice","seq":1,"service":"test","url":"test:","ucts":"2009-02-13T15:31:30.005-08:00"}]}]}`, body)

	// GET /search?q=alice@test
	req, err = http.NewRequest("GET", "/search?q=alice@test", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"results":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","users":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","name":"alice","seq":1,"service":"test","url":"test:","ucts":"2009-02-13T15:31:30.005-08:00"}]}]}`, body)

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
	require.Equal(t, `{"results":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","users":[{"kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","name":"alice","seq":1,"service":"test","url":"test:","ucts":"2009-02-13T15:31:30.005-08:00"}]}]}`, body)
}
