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

	alice := keys.NewEd25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	saveUser(t, env, alice, "alice", "github")

	err := env.client.Check(alice)
	require.NoError(t, err)
}
