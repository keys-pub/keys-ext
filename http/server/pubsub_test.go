package server_test

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keysd/http/api"
	"github.com/keys-pub/keysd/http/server"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
	env := newEnv(t)
	srv := newTestPubSubServer(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	charlie := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

	closeFn := srv.Start()
	defer closeFn()

	// GET /subscribe/:kid (alice)
	path := ds.Path("subscribe", alice.ID())
	conn := srv.WebsocketDial(t, path, clock, alice)
	defer conn.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	var b []byte
	var readErr error
	go func() {
		_, b, readErr = conn.ReadMessage()
		wg.Done()
	}()

	// POST /publish/:kid/:rid (charlie to alice)
	req, err := api.NewRequest("POST", ds.Path("publish", charlie.ID(), alice.ID()), bytes.NewReader([]byte("hi")), clock.Now(), charlie)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	wg.Wait()

	// Check read
	require.NoError(t, readErr)
	expected := `hi`
	require.Equal(t, expected, string(b))
}

func TestPubSubImpl(t *testing.T) {
	ps := server.NewPubSub()

	err := ps.Publish(context.TODO(), "topic1", []byte("ping1"))
	require.NoError(t, err)
	err = ps.Publish(context.TODO(), "topic1", []byte("ping2"))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vals := []string{}
	err = ps.Subscribe(ctx, "topic1", func(b []byte) {
		vals = append(vals, string(b))
		if len(vals) == 2 {
			cancel()
		}
	})
	require.NoError(t, err)

	require.Equal(t, []string{"ping1", "ping2"}, vals)
}

func TestWebsocket(t *testing.T) {
	env := newEnv(t)
	srv := newTestPubSubServer(t, env)
	clock := env.clock

	closeFn := srv.Start()
	defer closeFn()

	conn := srv.WebsocketDial(t, "/wsecho", clock, nil)
	defer conn.Close()

	err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
	require.NoError(t, err)

	_, b, err := conn.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, b, []byte("ping"))

	err = conn.WriteMessage(websocket.CloseMessage, nil)
	require.NoError(t, err)
}
