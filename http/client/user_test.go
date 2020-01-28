package client

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func saveUser(t *testing.T, env *env, key *keys.SignKey, name string, service string) *keys.Statement {
	url := ""
	switch service {
	case "github":
		url = fmt.Sprintf("https://gist.github.com/%s/1", name)
	case "twitter":
		url = fmt.Sprintf("https://twitter.com/%s/status/1", name)
	default:
		t.Fatal("unsupported service in test")
	}

	sc := keys.NewSigchain(key.PublicKey())
	user, err := keys.NewUser(env.users, key.ID(), service, name, url, sc.LastSeq()+1)
	require.NoError(t, err)
	st, err := keys.GenerateUserStatement(sc, user, key, env.clock.Now())
	require.NoError(t, err)

	msg, err := user.Sign(key)
	require.NoError(t, err)
	env.req.SetResponse(url, []byte(msg))

	err = env.client.PutSigchainStatement(st)
	require.NoError(t, err)

	// err = env.client.Check(key)
	// require.NoError(t, err)

	return st
}

func TestUserSearch(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	env := testEnv(t)
	defer env.closeFn()

	for i := 0; i < 10; i++ {
		key := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{byte(i)}, 32)))
		t.Logf("%s", key.ID())
		username := fmt.Sprintf("a%d", i)
		saveUser(t, env, key, username, "github")
	}

	resp, err := env.client.UserSearch("", 0)
	require.NoError(t, err)
	require.Equal(t, 10, len(resp.Results))
	require.NotNil(t, resp.Results[0].UserResult)
	require.Equal(t, "a0", resp.Results[0].UserResult.User.Name)

	resp, err = env.client.UserSearch("", 1)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Results))
	require.NotNil(t, resp.Results[0].UserResult)
	require.Equal(t, "a0", resp.Results[0].UserResult.User.Name)

	resp, err = env.client.UserSearch("a1", 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Results))
	require.NotNil(t, resp.Results[0].UserResult)
	require.Equal(t, "a1", resp.Results[0].UserResult.User.Name)

	resp, err = env.client.UserSearch("z", 1)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp.Results))
}
