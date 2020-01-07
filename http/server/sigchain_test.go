package server

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestSigchains(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	// firestore.SetContextLogger(NewContextLogger(DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	// fs := testFirestore(t)
	rq := keys.NewMockRequestor()
	users := testUserStore(t, fi, rq, clock)
	srv := newTestServer(t, clock, fi, users)
	// clock := newClockAtNow()
	// srv := newDevServer(t)

	alice, err := keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	require.NoError(t, err)
	aliceID := alice.ID()

	bob, err := keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	require.NoError(t, err)

	// GET /invalidloc (not found)
	req, err := http.NewRequest("GET", "/invalidloc", nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	expected := `{"error":{"code":404,"message":"resource not found"}}`
	require.Equal(t, expected, body)

	// PUT /sigchains (method not allowed)
	req, err = http.NewRequest("PUT", "/sigchains", bytes.NewReader([]byte("test")))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusMethodNotAllowed, code)
	expected = `{"error":{"code":405,"message":"method not allowed"}}`
	require.Equal(t, expected, body)

	// Alice sign "testing"
	aliceSpk := alice.PublicKey()
	aliceSc := keys.NewSigchain(aliceSpk)
	aliceSt, err := keys.GenerateStatement(aliceSc, []byte("testing"), alice, "", clock.Now())
	require.NoError(t, err)
	err = aliceSc.Add(aliceSt)
	require.NoError(t, err)
	aliceStBytes := aliceSt.Bytes()

	// PUT /sigchain/:kid/:seq
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(aliceStBytes))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "", body)

	// PUT /sigchain/:kid/:seq again (conflict error)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(aliceStBytes))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusConflict, code)
	expected = `{"error":{"code":409,"message":"statement already exists"}}`
	require.Equal(t, expected, body)

	// Bob sign "testing"
	bobSpk := bob.PublicKey()
	bobSc := keys.NewSigchain(bobSpk)
	bobSt, err := keys.GenerateStatement(bobSc, []byte("testing"), bob, "", clock.Now())
	require.NoError(t, err)

	// PUT /sigchain/:kid/:seq (invalid, bob's statement)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader([]byte(bobSt.Bytes())))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected = `{"error":{"code":400,"message":"invalid kid"}}`
	require.Equal(t, expected, body)

	// PUT /sigchain/:kid/:seq (empty json)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader([]byte("{}")))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected = `{"error":{"code":400,"message":"not enough bytes for statement"}}`
	require.Equal(t, expected, body)

	// PUT /sigchain/:kid/:seq (no body)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected = `{"error":{"code":400,"message":"missing body"}}`
	require.Equal(t, expected, body)

	// GET /sigchain/:kid/:seq
	req, err = http.NewRequest("GET", fmt.Sprintf("/sigchain/%s/1", alice.ID()), nil)
	require.NoError(t, err)
	code, header, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "Fri, 13 Feb 2009 15:31:30 GMT", header.Get("CreatedAt"))
	require.Equal(t, "2009-02-13T15:31:30.002-08:00", header.Get("CreatedAt-RFC3339M"))
	require.Equal(t, "Fri, 13 Feb 2009 15:31:30 GMT", header.Get("Last-Modified"))
	require.Equal(t, "2009-02-13T15:31:30.002-08:00", header.Get("Last-Modified-RFC3339M"))
	expectedSigned := `{".sig":"XFJAstMz0gv8ng4699WJx7m0UzLdbvzS27bpVq9Url+hiDNqKmmKcP0bGjmPsCzlNOu0o6lUrc0xXrmKX4WvCQ==","data":"dGVzdGluZw==","kid":"ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw","seq":1,"ts":1234567890001}`
	require.Equal(t, expectedSigned, body)

	// GET /sigchain/:kid
	req, err = http.NewRequest("GET", fmt.Sprintf("/sigchain/%s", alice.ID()), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedSigchain := `{"kid":"ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw","statements":[{".sig":"XFJAstMz0gv8ng4699WJx7m0UzLdbvzS27bpVq9Url+hiDNqKmmKcP0bGjmPsCzlNOu0o6lUrc0xXrmKX4WvCQ==","data":"dGVzdGluZw==","kid":"ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw","seq":1,"ts":1234567890001}]}`
	require.Equal(t, expectedSigchain, body)

	// GET /sigchain/:kid (not found)
	req, err = http.NewRequest("GET", keys.Path("sigchain", keys.RandID(keys.SignKeyType)), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"sigchain not found"}}`, body)

	// GET /sigchain/:kid?include=md
	req, err = http.NewRequest("GET", fmt.Sprintf("/sigchain/%s?include=md", alice.ID()), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedSigchain2 := `{"kid":"ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw","md":{"/ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw/1":{"createdAt":"2009-02-13T15:31:30.002-08:00","updatedAt":"2009-02-13T15:31:30.002-08:00"}},"statements":[{".sig":"XFJAstMz0gv8ng4699WJx7m0UzLdbvzS27bpVq9Url+hiDNqKmmKcP0bGjmPsCzlNOu0o6lUrc0xXrmKX4WvCQ==","data":"dGVzdGluZw==","kid":"ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw","seq":1,"ts":1234567890001}]}`
	require.Equal(t, expectedSigchain2, body)

	// GET /sigchains
	req, err = http.NewRequest("GET", "/sigchains", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedSigs := `{"statements":[{".sig":"XFJAstMz0gv8ng4699WJx7m0UzLdbvzS27bpVq9Url+hiDNqKmmKcP0bGjmPsCzlNOu0o6lUrc0xXrmKX4WvCQ==","data":"dGVzdGluZw==","kid":"ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw","seq":1,"ts":1234567890001}],"version":"1234567890003"}`
	require.Equal(t, expectedSigs, body)

	// GET /sigchains?include=md&limit=1
	req, err = http.NewRequest("GET", "/sigchains?include=md&limit=1", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedSigsWithMetadata := `{"md":{"/ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw/1":{"createdAt":"2009-02-13T15:31:30.002-08:00","updatedAt":"2009-02-13T15:31:30.002-08:00"}},"statements":[{".sig":"XFJAstMz0gv8ng4699WJx7m0UzLdbvzS27bpVq9Url+hiDNqKmmKcP0bGjmPsCzlNOu0o6lUrc0xXrmKX4WvCQ==","data":"dGVzdGluZw==","kid":"ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw","seq":1,"ts":1234567890001}],"version":"1234567890003"}`
	require.Equal(t, expectedSigsWithMetadata, body)

	// GET /:kid
	req, err = http.NewRequest("GET", keys.Path(aliceID), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedSigchain = `{"kid":"ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw","statements":[{".sig":"XFJAstMz0gv8ng4699WJx7m0UzLdbvzS27bpVq9Url+hiDNqKmmKcP0bGjmPsCzlNOu0o6lUrc0xXrmKX4WvCQ==","data":"dGVzdGluZw==","kid":"ed132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqrkl9gw","seq":1,"ts":1234567890001}]}`
	require.Equal(t, expectedSigchain, body)

	// Alice sign "testing2"
	aliceSt2, err := keys.GenerateStatement(aliceSc, []byte("testing2"), alice, "", clock.Now())
	require.NoError(t, err)
	err = aliceSc.Add(aliceSt2)
	require.NoError(t, err)

	// GET /:kid/:seq
	req, err = http.NewRequest("GET", fmt.Sprintf("/%s/1", alice.ID()), nil)
	require.NoError(t, err)
	code, header, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "Fri, 13 Feb 2009 15:31:30 GMT", header.Get("CreatedAt"))
	require.Equal(t, "2009-02-13T15:31:30.002-08:00", header.Get("CreatedAt-RFC3339M"))
	require.Equal(t, "Fri, 13 Feb 2009 15:31:30 GMT", header.Get("Last-Modified"))
	require.Equal(t, "2009-02-13T15:31:30.002-08:00", header.Get("Last-Modified-RFC3339M"))
	expectedSigned = `{".sig":"RQcZiGchACuPFiIIulcrfJ7d7Sb44EERqgxhlnZg4DFa6GstTY3dx0j+MaQVx42VcHm4E8Xi29CxrVZ+dcwyCg==","data":"dGVzdGluZw==","kid":"HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec","seq":1,"ts":1234567890001}`

	// PUT /:kid/:seq
	req, err = http.NewRequest("PUT", fmt.Sprintf("/%s/2", alice.ID()), bytes.NewReader(aliceSt2.Bytes()))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "", body)

	// PUT /invalidloc/:seq
	req, err = http.NewRequest("PUT", keys.Path("invalidloc", 1), bytes.NewReader(aliceSt2.Bytes()))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"invalid kid"}}`, body)
}
