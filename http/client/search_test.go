package client

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func saveUser(t *testing.T, env *env, key keys.Key, name string, service string) *keys.Statement {
	url := ""
	switch service {
	case "github":
		url = fmt.Sprintf("https://gist.github.com/%s/1", name)
	case "twitter":
		url = fmt.Sprintf("https://twitter.com/%s/status/1", name)
	default:
		t.Fatal("unsupported service in test")
	}

	sc := keys.NewSigchain(key.SignKey().PublicKey)
	user, err := keys.NewUser(env.uc, key.ID(), service, name, url, sc.LastSeq()+1)
	require.NoError(t, err)
	st, err := keys.GenerateUserStatement(sc, user, key.SignKey(), env.clock.Now())
	require.NoError(t, err)

	msg, err := user.Sign(key.SignKey())
	require.NoError(t, err)
	env.req.SetResponse(url, []byte(msg))

	err = env.client.PutSigchainStatement(st)
	require.NoError(t, err)

	// err = env.client.Check(key)
	// require.NoError(t, err)

	return st
}

func TestSearch(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	env := testEnv(t)
	defer env.closeFn()

	for i := 0; i < 10; i++ {
		key, err := keys.NewKey(keys.Bytes32(bytes.Repeat([]byte{byte(i)}, 32)))
		require.NoError(t, err)
		username := fmt.Sprintf("a%d", i)
		saveUser(t, env, key, username, "github")
	}

	resp, err := env.client.Search("", 0, 0)
	require.NoError(t, err)
	require.Equal(t, 10, len(resp.Results))
	require.Equal(t, 1, len(resp.Results[0].Users))
	require.Equal(t, "a0", resp.Results[0].Users[0].User.Name)

	resp, err = env.client.Search("", 0, 1)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Results))
	require.Equal(t, 1, len(resp.Results[0].Users))
	require.Equal(t, "a0", resp.Results[0].Users[0].User.Name)

	resp, err = env.client.Search("a1", 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Results))
	require.Equal(t, 1, len(resp.Results[0].Users))
	require.Equal(t, "a1", resp.Results[0].Users[0].User.Name)

	resp, err = env.client.Search("z", 0, 1)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp.Results))

	resp, err = env.client.Search("ddRZXkYg1VcHRh", 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Results))
	require.Equal(t, "ddRZXkYg1VcHRhpR6zu5kPBzsSLV9sJTWkTdduCJu2yu", resp.Results[0].KID.String())
}
