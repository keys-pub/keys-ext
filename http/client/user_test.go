package client

import (
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	env := testEnv(t)
	defer env.closeFn()

	key, err := keys.NewKeyFromSeedPhrase(aliceSeed, true)
	require.NoError(t, err)
	saveUser(t, env, key, "alice", "github")

	err = env.client.Check(key)
	require.NoError(t, err)
}
