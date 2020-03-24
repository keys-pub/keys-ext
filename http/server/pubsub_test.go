package server_test

import (
	"bytes"
	"net/http"
	"sync"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
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
	req, err := api.NewRequest("POST", keys.Path("publish", charlie.ID(), alice.ID()), bytes.NewReader([]byte("hi")), clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	wg.Wait()

	// Check read
	require.NoError(t, readErr)
	expected := `hi`
	require.Equal(t, expected, string(b[:n]))
}
