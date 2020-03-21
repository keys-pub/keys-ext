package server_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/stretchr/testify/require"
)

func TestEphem(t *testing.T) {
	// server.SetContextLogger(server.NewContextLogger(server.DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	users := testUserStore(t, fi, rq, clock)
	srv := newTestServer(t, clock, fi, users)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	// PUT /ephem/:kid/:rid/:id
	req, err := api.NewRequest("PUT", keys.Path("ephem", alice.ID(), charlie.ID(), "offer"), bytes.NewReader([]byte("hi")), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /ephem/:kid/:rid/:id
	req, err = api.NewRequest("GET", keys.Path("ephem", alice.ID(), charlie.ID(), "offer"), nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "hi", body)
}
