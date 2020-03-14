package client

import (
	"bytes"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	env := testEnv(t)
	defer env.closeFn()
	ks := keys.NewMemKeystore()
	client := testClient(t, env, ks)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	saveUser(t, env, client, alice, "alice", "github")

	err := client.Check(alice)
	require.NoError(t, err)
}
