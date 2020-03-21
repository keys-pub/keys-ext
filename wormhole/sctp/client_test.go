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

	err = alice.Write([]byte("ping"))
	require.NoError(t, err)

	bobWg.Wait()

	buf := make([]byte, 1024)
	n, err := bob.Read(buf)
	require.NoError(t, err)
	require.Equal(t, "ping", string(buf[:n]))

	err = bob.Write([]byte("ping"))
	require.NoError(t, err)
	n, err = alice.Read(buf)
	require.NoError(t, err)
	require.Equal(t, "ping", string(buf[:n]))

	alice.Close()
	bob.Close()
}
