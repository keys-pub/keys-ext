package client_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	env, closeFn := newEnv(t, server.NoLevel)
	defer closeFn()
	client := newTestClient(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	saveUser(t, env, client, alice, "alice", "github")

	err := client.Check(context.TODO(), alice)
	require.NoError(t, err)
}
