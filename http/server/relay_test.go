package server_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/server"
	"github.com/stretchr/testify/require"
)

func TestRelay(t *testing.T) {
	server.SetContextLogger(server.NewContextLogger(server.DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	srv := newTestServer(t, clock, fi, nil)

	// POST /relay/:id
	req, err := http.NewRequest("POST", keys.Path("relay", "offer"), bytes.NewReader([]byte("hi")))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// PUT /relay/:id
	req, err = http.NewRequest("PUT", keys.Path("relay", "offer"), bytes.NewReader([]byte("hi")))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// GET /relay/:id
	req, err = http.NewRequest("GET", keys.Path("relay", "offer"), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "hi", body)

	// GET /relay/:id (again)
	req, err = http.NewRequest("GET", keys.Path("relay", "offer"), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, body)
}
