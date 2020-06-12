package wormhole_test

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"

	"github.com/keys-pub/keys-ext/wormhole"
	"github.com/keys-pub/keys-ext/wormhole/sctp"
	"github.com/stretchr/testify/require"
)

// TODO: SCTP write buffer?
// TODO: Keep alive?
// TODO: Close, reconnect?
// TODO: Messages could have been omitted by network, include previous message ID

func testKeyring(t *testing.T) *keyring.Keyring {
	kr, err := keyring.New(keyring.Mem())
	require.NoError(t, err)
	kr.SetMasterKey(keys.Rand32())
	return kr
}

func TestNew(t *testing.T) {
	// wormhole.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))
	// sctp.SetLogger(sctp.NewLogger(sctp.DebugLevel))
	var err error
	env := testEnv(t)
	defer env.closeFn()

	// Local
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	kra := testKeyring(t)
	err = keys.Save(kra, alice)
	require.NoError(t, err)
	err = keys.Save(kra, bob.PublicKey())
	require.NoError(t, err)

	krb := testKeyring(t)
	err = keys.Save(krb, bob)
	require.NoError(t, err)
	err = keys.Save(krb, alice.PublicKey())
	require.NoError(t, err)

	testWormhole(t, env, true, alice, bob, kra, krb)

	// Remote
	// testWormhole(t, env, false)
}

func TestWormholeSameKey(t *testing.T) {
	// wormhole.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))
	// sctp.SetLogger(sctp.NewLogger(sctp.DebugLevel))
	var err error
	env := testEnv(t)
	defer env.closeFn()

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	kra := testKeyring(t)
	err = keys.Save(kra, alice)
	require.NoError(t, err)

	testWormhole(t, env, true, alice, alice, kra, kra)
}

func testWormhole(t *testing.T, env *env, local bool, alice *keys.EdX25519Key, bob *keys.EdX25519Key, kra *keyring.Keyring, krb *keyring.Keyring) {
	ctx := context.TODO()

	openWg := &sync.WaitGroup{}
	openWg.Add(2)

	closeWg := &sync.WaitGroup{}
	closeWg.Add(2)

	server := env.httpServer.URL
	wha, err := wormhole.New(server, kra)
	require.NoError(t, err)
	defer wha.Close()
	wha.SetTimeNow(env.clock.Now)
	wha.OnStatus(func(st wormhole.Status) {
		switch st {
		case wormhole.Connected:
			openWg.Done()
		case wormhole.Closed:
			closeWg.Done()
		}
	})

	var offer *sctp.Addr
	if local {
		o, err := wha.CreateLocalOffer(ctx, alice.ID(), bob.ID())
		require.NoError(t, err)
		offer = o
	} else {
		o, err := wha.CreateOffer(ctx, alice.ID(), bob.ID())
		require.NoError(t, err)
		offer = o
	}

	inviteCode, err := wha.CreateInvite(ctx, alice.ID(), bob.ID())
	require.NoError(t, err)

	go func() {
		if err := wha.Connect(ctx, alice.ID(), bob.ID(), offer); err != nil {
			panic(err)
		}
	}()

	whb, err := wormhole.New(server, krb)
	require.NoError(t, err)
	defer whb.Close()
	whb.SetTimeNow(env.clock.Now)
	whb.OnStatus(func(st wormhole.Status) {
		switch st {
		case wormhole.Connected:
			openWg.Done()
		case wormhole.Closed:
			closeWg.Done()
		}
	})

	if inviteCode != "" {
		invite, err := whb.FindInvite(ctx, inviteCode)
		if err != nil {
			return
		}
		require.Equal(t, invite.Sender, alice.ID())
		require.Equal(t, invite.Recipient, bob.ID())
	}

	go func() {
		if err := whb.Listen(ctx, bob.ID(), alice.ID(), offer); err != nil {
			panic(err)
		}
	}()

	openWg.Wait()

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
	id := wormhole.NewID()
	pending, err := wha.WriteMessage(ctx, id, []byte("ping"), wormhole.UTF8Content)
	require.NoError(t, err)
	require.Equal(t, wormhole.Pending, pending.Type)
	require.Equal(t, id, pending.ID)

	msg, err := whb.ReadMessage(ctx, true)
	require.NoError(t, err)
	require.Equal(t, "ping", string(msg.Content.Data))
	require.Equal(t, id, string(msg.ID))

	reply, err := wha.ReadMessage(ctx, true)
	require.NoError(t, err)
	require.Equal(t, wormhole.Ack, reply.Type)
	require.Equal(t, id, reply.ID)

	wha.Close()

	_, err = whb.ReadMessage(ctx, true)
	require.EqualError(t, err, "closed")

	closeWg.Wait()
}

func TestWormholeCancel(t *testing.T) {
	// wormhole.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))
	// sctp.SetLogger(sctp.NewLogger(sctp.DebugLevel))

	env := testEnv(t)
	defer env.closeFn()

	testWormholeCancel(t, env, 100*time.Millisecond)
	testWormholeCancel(t, env, time.Second)
	// testWormholeCancel(t, env, time.Second*5)
}

func testWormholeCancel(t *testing.T, env *env, dt time.Duration) {
	var err error
	server := env.httpServer.URL

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	kra := testKeyring(t)
	err = keys.Save(kra, alice)
	require.NoError(t, err)

	wha, err := wormhole.New(server, kra)
	require.NoError(t, err)
	defer wha.Close()
	wha.SetTimeNow(env.clock.Now)
	ctx, cancel := context.WithTimeout(context.Background(), dt)
	defer cancel()

	offer := &sctp.Addr{IP: "127.0.0.1", Port: 1234}
	err = wha.Listen(ctx, alice.ID(), bob.ID(), offer)
	require.True(t, strings.HasSuffix(err.Error(), "context deadline exceeded"))

	// TODO: Test cancel with Connect
}

func TestWormholeNoRecipient(t *testing.T) {
	// wormhole.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))
	// sctp.SetLogger(sctp.NewLogger(sctp.DebugLevel))
	var err error
	env := testEnv(t)
	defer env.closeFn()
	server := env.httpServer.URL

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	kra := testKeyring(t)
	err = keys.Save(kra, alice)
	require.NoError(t, err)

	krb := testKeyring(t)
	err = keys.Save(krb, bob)
	require.NoError(t, err)

	wha, err := wormhole.New(server, kra)
	require.NoError(t, err)
	defer wha.Close()
	wha.SetTimeNow(env.clock.Now)

	whb, err := wormhole.New(server, krb)
	require.NoError(t, err)
	defer wha.Close()
	whb.SetTimeNow(env.clock.Now)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	offer, err := wha.CreateOffer(ctx, alice.ID(), bob.ID())
	require.NoError(t, err)
	// Don't Connect

	err = whb.Listen(ctx, alice.ID(), bob.ID(), offer)
	require.EqualError(t, err, "not found kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077")

	wha.Close()
}
