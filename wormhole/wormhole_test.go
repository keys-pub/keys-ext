package wormhole_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/keys-pub/keys"

	"github.com/keys-pub/keysd/wormhole"
	"github.com/stretchr/testify/require"
)

func TestNewWormhole(t *testing.T) {
	// wormhole.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))
	// sctp.SetLogger(sctp.NewLogger(sctp.DebugLevel))

	env := testEnv(t)
	defer env.closeFn()
	server := env.httpServer.URL

	alice := keys.GenerateEdX25519Key()
	bob := keys.GenerateEdX25519Key()

	ksa := keys.NewMemKeystore()
	err := ksa.SaveEdX25519Key(alice)
	require.NoError(t, err)

	ksb := keys.NewMemKeystore()
	err = ksb.SaveEdX25519Key(bob)
	require.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	wha, err := wormhole.NewWormhole(server, ksa)
	require.NoError(t, err)
	defer wha.Close()
	wha.SetTimeNow(env.clock.Now)
	wha.OnConnect(func() {
		wg.Done()
	})
	go func() {
		err = wha.Start(context.TODO(), alice, bob.PublicKey())
		if err != nil {
			panic(err)
		}
	}()

	whb, err := wormhole.NewWormhole(server, ksb)
	require.NoError(t, err)
	defer whb.Close()
	whb.SetTimeNow(env.clock.Now)
	whb.OnConnect(func() {
		wg.Done()
	})
	go func() {
		err = whb.Start(context.TODO(), bob, alice.PublicKey())
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()

	// Send ping/pong
	msgWg := sync.WaitGroup{}
	msgWg.Add(1)

	whb.OnMessage(func(data []byte) {
		if string(data) == "ping" {
			err := whb.Send([]byte("pong"))
			require.NoError(t, err)
		}
	})

	wha.OnMessage(func(data []byte) {
		msgWg.Done()
	})

	err = wha.Send([]byte("ping"))
	require.NoError(t, err)

	msgWg.Wait()

	// Close
	closeWg := &sync.WaitGroup{}
	closeWg.Add(2)
	wha.OnClose(func() {
		closeWg.Done()
	})
	wha.OnClose(func() {
		closeWg.Done()
	})

	wha.Close()
	whb.Close()
}

func TestWormholeCancel(t *testing.T) {
	// wormhole.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))
	// webrtc.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))

	env := testEnv(t)
	defer env.closeFn()

	testWormholeCancel(t, env, 100*time.Millisecond)
	testWormholeCancel(t, env, time.Second)
	// testWormholeCancel(t, env, time.Second*5)
}

func testWormholeCancel(t *testing.T, env *env, dt time.Duration) {
	server := env.httpServer.URL

	alice := keys.GenerateEdX25519Key()
	bob := keys.GenerateEdX25519Key()

	ksa := keys.NewMemKeystore()
	err := ksa.SaveEdX25519Key(alice)
	require.NoError(t, err)

	wha, err := wormhole.NewWormhole(server, ksa)
	require.NoError(t, err)
	defer wha.Close()
	wha.SetTimeNow(env.clock.Now)
	ctx, cancel := context.WithTimeout(context.Background(), dt)
	defer cancel()
	err = wha.Start(ctx, alice, bob.PublicKey())
	require.True(t, strings.HasSuffix(err.Error(), "context deadline exceeded"))
}
