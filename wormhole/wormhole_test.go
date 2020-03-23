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

// TODO: SCTP write buffer?
// TODO: Keep alive?
// TODO: Close, reconnect?
// TODO: Messages could have been omitted by network, include previous message ID

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

	ctx := context.TODO()

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
		err = wha.Start(ctx, alice, bob.PublicKey())
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
		err = whb.Start(ctx, bob, alice.PublicKey())
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()

	err = wha.Write(ctx, []byte("ping"))
	require.NoError(t, err)

	go func() {
		b, err := whb.Read(ctx)
		require.NoError(t, err)
		require.Equal(t, "ping", string(b))
		err = whb.Write(ctx, []byte("pong"))
		require.NoError(t, err)
	}()

	b, err := wha.Read(ctx)
	require.NoError(t, err)
	require.Equal(t, "pong", string(b))

	// Message
	pending, err := wha.WriteMessage(ctx, []byte("ping"), wormhole.UTF8Content)
	require.NoError(t, err)
	require.Equal(t, pending.Type, wormhole.Pending)

	msg, err := whb.ReadMessage(ctx, true)
	require.NoError(t, err)
	require.Equal(t, "ping", string(msg.Content.Data))
	require.Equal(t, pending.ID, string(msg.ID))

	reply, err := wha.ReadMessage(ctx, true)
	require.NoError(t, err)
	require.Equal(t, wormhole.Ack, reply.Type)
	require.Equal(t, pending.ID, reply.ID)

	// Close
	closeWg := &sync.WaitGroup{}
	closeWg.Add(2)
	wha.OnClose(func() {
		closeWg.Done()
	})
	whb.OnClose(func() {
		closeWg.Done()
	})

	wha.Close()

	_, err = whb.ReadMessage(ctx, true)
	require.EqualError(t, err, "closed")

	closeWg.Wait()
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

func TestWormholeNoRecipient(t *testing.T) {
	wormhole.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))
	// sctp.SetLogger(sctp.NewLogger(sctp.DebugLevel))

	env := testEnv(t)
	defer env.closeFn()
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
	wha.OnConnect(func() {
		t.Fatalf("Should timeout")
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = wha.Start(ctx, alice, bob.PublicKey())
	require.EqualError(t, err, "context deadline exceeded")

	wha.Close()
}
