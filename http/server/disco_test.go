package server_test

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys-ext/http/api"
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
	req, err := api.NewRequest("PUT", ds.Path("disco", alice.ID(), charlie.ID(), "offer")+"?expire=1m", bytes.NewReader([]byte("hi")), env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /disco/:kid/:rid/offer (charlie from alice)
	req, err = api.NewRequest("GET", ds.Path("disco", alice.ID(), charlie.ID(), "offer"), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `hi`, body)

	// GET (again)
	req, err = api.NewRequest("GET", ds.Path("disco", alice.ID(), charlie.ID(), "offer"), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /disco/:kid/:rid/offer (alice to charlie, 1m)
	req, err = api.NewRequest("PUT", ds.Path("disco", alice.ID(), charlie.ID(), "offer")+"?expire=1m", bytes.NewReader([]byte("hi")), env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// DEL (invalid auth)
	req, err = api.NewRequest("DELETE", ds.Path("disco", alice.ID(), charlie.ID()), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// DEL /disco/:kid/:rid
	req, err = api.NewRequest("DELETE", ds.Path("disco", alice.ID(), charlie.ID()), nil, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)

	// GET (charlie, after delete)
	req, err = api.NewRequest("GET", ds.Path("disco", alice.ID(), charlie.ID(), "offer"), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /disco/:kid/:rid/offer (expire 1ms)
	req, err = api.NewRequest("PUT", ds.Path("disco", alice.ID(), charlie.ID(), "offer")+"?expire=1ms", bytes.NewReader([]byte("hi")), env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)
	time.Sleep(time.Millisecond)

	// GET (after expire)
	req, err = api.NewRequest("GET", ds.Path("disco", alice.ID(), charlie.ID(), "offer"), nil, env.clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /disco/:kid/:rid/offer (alice to alice, 1m)
	req, err = api.NewRequest("PUT", ds.Path("disco", alice.ID(), alice.ID(), "offer")+"?expire=1m", bytes.NewReader([]byte("hi")), env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /disco/:kid/:rid/offer (alice to alice)
	req, err = api.NewRequest("GET", ds.Path("disco", alice.ID(), alice.ID(), "offer"), nil, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `hi`, body)

	// DEL /disco/:kid/:rid (alice to alice)
	req, err = api.NewRequest("DELETE", ds.Path("disco", alice.ID(), alice.ID()), nil, env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)
}
