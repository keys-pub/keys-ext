package server_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/stretchr/testify/require"
)

func TestShare(t *testing.T) {
	// api.SetLogger(api.NewLogger(api.DebugLevel))

	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)

	key := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	key2 := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	// PUT /share/:kid (1m)
	content := []byte("test1")
	contentHash := http.ContentHash(content)
	req, err := http.NewAuthRequest("PUT", dstore.Path("share", key.ID())+"?expire=1m", bytes.NewReader(content), contentHash, env.clock.Now(), http.Authorization(key))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, `{}`, body)
	require.Equal(t, http.StatusOK, code)

	// GET /share/:kid (bad key)
	req, err = http.NewAuthRequest("GET", dstore.Path("share", key.ID()), nil, "", env.clock.Now(), http.Authorization(key2))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)

	// GET /share/:kid
	req, err = http.NewAuthRequest("GET", dstore.Path("share", key.ID()), nil, "", env.clock.Now(), http.Authorization(key))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, string(content), body)

	// GET (again)
	req, err = http.NewAuthRequest("GET", dstore.Path("share", key.ID()), nil, "", env.clock.Now(), http.Authorization(key))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /share/:kid (expire 1ms)
	req, err = http.NewAuthRequest("PUT", dstore.Path("share", key.ID())+"?expire=1ms", bytes.NewReader(content), contentHash, env.clock.Now(), http.Authorization(key))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, body)
	time.Sleep(time.Millisecond)

	// GET (after expire)
	req, err = http.NewAuthRequest("GET", dstore.Path("share", key.ID()), nil, "", env.clock.Now(), http.Authorization(key))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)

	// PUT /share/:kid (30m)
	req, err = http.NewAuthRequest("PUT", dstore.Path("share", key.ID())+"?expire=30m", bytes.NewReader(content), contentHash, env.clock.Now(), http.Authorization(key))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"max expire is 15m"}}`, body)

	// PUT /share/:kid (bad key)
	req, err = http.NewAuthRequest("PUT", dstore.Path("share", key.ID())+"?expire=30m", bytes.NewReader(content), contentHash, env.clock.Now(), http.Authorization(key2))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)
}
