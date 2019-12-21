package server

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/stretchr/testify/require"
)

func TestShare(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	users := keys.NewTestUserStore(fi, keys.NewSigchainStore(fi), rq, clock.Now)
	srv := newTestServer(t, clock, fi, users)

	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)

	bob, err := keys.NewKeyFromSeedPhrase(bobSeed, false)
	require.NoError(t, err)

	group, err := keys.NewKeyFromSeedPhrase(groupSeed, false)
	require.NoError(t, err)

	// PUT /share/:kid/:recipient (alice)
	req, err := api.NewRequest("PUT", keys.Path("share", alice.ID(), group.ID()), bytes.NewReader([]byte("ok")), clock.Now(), group)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// PUT /share/:kid/:recipient (bob invalid)
	req, err = api.NewRequest("PUT", keys.Path("share", alice.ID(), group.ID()), bytes.NewReader([]byte("ok")), clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// GET /share/:kid/:recipient (alice)
	req, err = api.NewRequest("GET", keys.Path("share", alice.ID(), group.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "ok", body)

	// GET /share/:kid/:recipient (bob, not found)
	req, err = api.NewRequest("GET", keys.Path("share", bob.ID(), group.ID()), nil, clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)

	// GET /share/:kid/:recipient (bob invalid)
	req, err = api.NewRequest("GET", keys.Path("share", alice.ID(), group.ID()), nil, clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, body)

	// PUT /share/:kid/:recipient (bob)
	req, err = api.NewRequest("PUT", keys.Path("share", bob.ID(), group.ID()), bytes.NewReader([]byte("ok bob")), clock.Now(), group)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /share/:kid/:recipient (bob)
	req, err = api.NewRequest("GET", keys.Path("share", bob.ID(), group.ID()), nil, clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "ok bob", body)

	// PUT /share/:kid/:recipient (alice, update)
	req, err = api.NewRequest("PUT", keys.Path("share", alice.ID(), group.ID()), bytes.NewReader([]byte("ok2")), clock.Now(), group)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /share/:kid/:recipient (alice)
	req, err = api.NewRequest("GET", keys.Path("share", alice.ID(), group.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "ok2", body)

	// DELETE /share/:kid/:recipient (alice, delete)
	req, err = api.NewRequest("DELETE", keys.Path("share", alice.ID(), group.ID()), nil, clock.Now(), group)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /share/:kid/:recipient (alice, not found after delete)
	req, err = api.NewRequest("GET", keys.Path("share", alice.ID(), group.ID()), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
}
