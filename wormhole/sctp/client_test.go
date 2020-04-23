package sctp_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/keys-pub/keysd/wormhole/sctp"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	// sctp.SetLogger(sctp.NewLogger(sctp.DebugLevel))
	// testClient(t, true)
	testClient(t, false)
}

func testClient(t *testing.T, stun bool) {
	alice := sctp.NewClient()
	bob := sctp.NewClient()
	defer alice.Close()
	defer bob.Close()

	ctx := context.TODO()

	var aliceAddr *sctp.Addr
	var bobAddr *sctp.Addr
	if stun {
		a, err := alice.STUN(ctx, time.Second*5)
		require.NoError(t, err)
		aliceAddr = a
		b, err := bob.STUN(ctx, time.Second*5)
		require.NoError(t, err)
		bobAddr = b
	} else {
		a, err := alice.Local()
		require.NoError(t, err)
		aliceAddr = a
		b, err := bob.Local()
		require.NoError(t, err)
		bobAddr = b
	}

	aliceWg := &sync.WaitGroup{}
	aliceWg.Add(1)

	go func() {
		err := alice.Connect(ctx, bobAddr)
		require.NoError(t, err)
		aliceWg.Done()
	}()

	bobWg := &sync.WaitGroup{}
	bobWg.Add(1)

	go func() {
		err := bob.ListenForPeer(ctx, aliceAddr)
		require.NoError(t, err)
		bobWg.Done()
	}()

	aliceWg.Wait()

	err := alice.Write(ctx, []byte("ping"))
	require.NoError(t, err)

	bobWg.Wait()

	buf := make([]byte, 1024)
	n, err := bob.Read(ctx, buf)
	require.NoError(t, err)
	require.Equal(t, "ping", string(buf[:n]))

	err = bob.Write(ctx, []byte("ping"))
	require.NoError(t, err)
	n, err = alice.Read(ctx, buf)
	require.NoError(t, err)
	require.Equal(t, "ping", string(buf[:n]))

	// Read timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = alice.Read(ctx, buf)
	require.EqualError(t, err, "context deadline exceeded")

	alice.Close()
	bob.Close()
}

func TestParseAddr(t *testing.T) {
	a, err := sctp.ParseAddr("1.2.3.4:123")
	require.NoError(t, err)
	require.Equal(t, "1.2.3.4", a.IP)
	require.Equal(t, 123, a.Port)
}
