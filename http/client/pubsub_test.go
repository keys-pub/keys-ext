package client

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t)
	defer env.closeFn()

	ksa := keys.NewMemKeystore()
	aliceClient := testClient(t, env, ksa)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	err := ksa.SaveEdX25519Key(alice)
	require.NoError(t, err)

	ksb := keys.NewMemKeystore()
	bobClient := testClient(t, env, ksb)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	err = ksb.SaveEdX25519Key(bob)
	require.NoError(t, err)

	// Pub
	err = aliceClient.Publish(context.TODO(), alice.ID(), bob.ID(), []byte("hi"))
	require.NoError(t, err)

	// Sub
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	msgs := []string{}
	receiveFn := func(msg *PubSubMessage) {
		msgs = append(msgs, string(msg.Data))
		require.Equal(t, alice.ID(), msg.KID)
		t.Logf("msg: %v", msg)
		if len(msgs) >= 2 {
			cancel()
		}
	}
	go func() {
		err := bobClient.Subscribe(ctx, bob.ID(), receiveFn)
		require.NoError(t, err)
		wg.Done()
	}()

	// Pub
	err = aliceClient.Publish(context.TODO(), alice.ID(), bob.ID(), []byte("what time is the meeting?"))
	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, []string{"hi", "what time is the meeting?"}, msgs)

}
