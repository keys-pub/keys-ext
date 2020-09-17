package client_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestShare(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	// logger = NewLogger(DebugLevel)

	env, closeFn := newEnv(t)
	defer closeFn()

	client := newTestClient(t, env)
	key := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// Put
	err := client.ShareSeal(context.TODO(), key, []byte("hi"), time.Minute)
	require.NoError(t, err)

	// Get
	out, err := client.ShareOpen(context.TODO(), key)
	require.NoError(t, err)
	require.Equal(t, "hi", string(out))

	// Get (again)
	out, err = client.ShareOpen(context.TODO(), key)
	require.NoError(t, err)
	require.Empty(t, out)
}
