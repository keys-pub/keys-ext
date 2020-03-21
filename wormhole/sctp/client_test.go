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

	alice := sctp.NewClient()
	bob := sctp.NewClient()
	defer alice.Close()
	defer bob.Close()

	aliceAddr, err := alice.STUN(context.TODO(), time.Second*5)
	require.NoError(t, err)
	bobAddr, err := bob.STUN(context.TODO(), time.Second*5)
	require.NoError(t, err)

	aliceWg := &sync.WaitGroup{}
	aliceWg.Add(1)

	go func() {
		err = alice.Connect(context.TODO(), bobAddr)
		require.NoError(t, err)
		aliceWg.Done()
	}()

	bobWg := &sync.WaitGroup{}
	bobWg.Add(1)

	go func() {
		err = bob.Listen(context.TODO(), aliceAddr)
		require.NoError(t, err)
		bobWg.Done()
	}()

	aliceWg.Wait()

	err = alice.Write(context.TODO(), []byte("ping"))
	require.NoError(t, err)

	bobWg.Wait()

	buf := make([]byte, 1024)
	n, err := bob.Read(context.TODO(), buf)
	require.NoError(t, err)
	require.Equal(t, "ping", string(buf[:n]))

	err = bob.Write(context.TODO(), []byte("ping"))
	require.NoError(t, err)
	n, err = alice.Read(context.TODO(), buf)
	require.NoError(t, err)
	require.Equal(t, "ping", string(buf[:n]))

	// Read timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	n, err = alice.Read(ctx, buf)
	require.EqualError(t, err, "stream read error: context deadline exceeded")

	alice.Close()
	bob.Close()
}
