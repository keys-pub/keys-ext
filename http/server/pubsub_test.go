package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/stretchr/testify/require"
)

func TestPubsub(t *testing.T) {
	// server.SetContextLogger(server.NewContextLogger(server.DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	users := testUserStore(t, fi, rq, clock)
	srv := newTestServer(t, clock, fi, users)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	closeFn := srv.Start()
	defer closeFn()

	// GET /subscribe/:kid (alice)
	path := keys.Path("subscribe", alice.ID())
	conn := srv.WebsocketDial(t, path, clock, alice)
	defer conn.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	var b []byte
	var n int
	var readErr error
	go func() {
		b = make([]byte, 1024)
		n, readErr = conn.Read(b)
		wg.Done()
	}()

	// POST /publish/:kid/:rid (charlie to alice)
	st := keys.NewSignedStatement([]byte("hi"), charlie, "", time.Time{})
	stb, err := json.Marshal(st)
	require.NoError(t, err)
	req, err := api.NewRequest("POST", keys.Path("publish", charlie.ID(), alice.ID()), bytes.NewReader(stb), clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	wg.Wait()

	// Check read
	require.NoError(t, readErr)
	expected := `{".sig":"X6cbvVKN5AUbjONPHqm6fn3voZgAVi9c77y04wRVhQdJOwS/9fBc77cS4171leT6lfJebEwd3oU5k1gURBVxCw==","data":"aGk=","kid":"kex1a4yj333g68pvd6hfqvufqkv4vy54jfe6t33ljd3kc9rpfty8xlgs2u3qxr"}`
	require.Equal(t, expected, string(b[:n]))

	// POST /publish/:kid/:rid (statement kid mismatch)
	st = keys.NewSignedStatement([]byte("hi"), alice, "", time.Time{})
	stb, err = json.Marshal(st)
	require.NoError(t, err)
	req, err = api.NewRequest("POST", keys.Path("publish", charlie.ID(), alice.ID()), bytes.NewReader(stb), clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"statement kid mismatch"}}`, body)
}
