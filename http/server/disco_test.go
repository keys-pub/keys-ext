package server_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/http"
	"github.com/stretchr/testify/require"
)

func TestDisco(t *testing.T) {
	// api.SetLogger(api.NewLogger(api.DebugLevel))

	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	// PUT /disco/:kid/:rid/offer (alice to charlie, 1m)
	content := []byte("testdata")
	contentHash := http.ContentHash(content)
	req, err := http.NewAuthRequest("PUT", docs.Path("disco", alice.ID(), charlie.ID(), "offer")+"?expire=1m", bytes.NewReader(content), contentHash, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /disco/:kid/:rid/offer (charlie from alice)
	req, err = http.NewAuthRequest("GET", docs.Path("disco", alice.ID(), charlie.ID(), "offer"), nil, "", env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, string(content), body)

	// GET (again)
	req, err = http.NewAuthRequest("GET", docs.Path("disco", alice.ID(), charlie.ID(), "offer"), nil, "", env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /disco/:kid/:rid/offer (alice to charlie, 1m)
	req, err = http.NewAuthRequest("PUT", docs.Path("disco", alice.ID(), charlie.ID(), "offer")+"?expire=1m", bytes.NewReader(content), contentHash, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// DEL (invalid auth)
	req, err = http.NewAuthRequest("DELETE", docs.Path("disco", alice.ID(), charlie.ID()), nil, "", env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// DEL /disco/:kid/:rid
	req, err = http.NewAuthRequest("DELETE", docs.Path("disco", alice.ID(), charlie.ID()), nil, "", env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET (charlie, after delete)
	req, err = http.NewAuthRequest("GET", docs.Path("disco", alice.ID(), charlie.ID(), "offer"), nil, "", env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /disco/:kid/:rid/offer (expire 1ms)
	req, err = http.NewAuthRequest("PUT", docs.Path("disco", alice.ID(), charlie.ID(), "offer")+"?expire=1ms", bytes.NewReader(content), contentHash, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)
	time.Sleep(time.Millisecond)

	// GET (after expire)
	req, err = http.NewAuthRequest("GET", docs.Path("disco", alice.ID(), charlie.ID(), "offer"), nil, "", env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /disco/:kid/:rid/offer (alice to alice, 1m)
	req, err = http.NewAuthRequest("PUT", docs.Path("disco", alice.ID(), alice.ID(), "offer")+"?expire=1m", bytes.NewReader(content), contentHash, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /disco/:kid/:rid/offer (alice to alice)
	req, err = http.NewAuthRequest("GET", docs.Path("disco", alice.ID(), alice.ID(), "offer"), nil, "", env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, string(content), body)

	// DEL /disco/:kid/:rid (alice to alice)
	req, err = http.NewAuthRequest("DELETE", docs.Path("disco", alice.ID(), alice.ID()), nil, "", env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)
}
