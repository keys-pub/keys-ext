package client_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	env := testEnv(t, nil)
	defer env.closeFn()
	client := testClient(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	saveUser(t, env, client, alice, "alice", "github")

	err := client.Check(context.TODO(), alice)
	require.NoError(t, err)
}
