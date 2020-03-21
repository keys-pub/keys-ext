package client

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func saveUser(t *testing.T, env *env, client *Client, key *keys.EdX25519Key, name string, service string) *keys.Statement {
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

	err = client.PutSigchainStatement(context.TODO(), st)
	require.NoError(t, err)

	// err = client.Check(key)
	// require.NoError(t, err)

	return st
}

func TestUserSearch(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	env := testEnv(t)
	defer env.closeFn()

	ks := keys.NewMemKeystore()
	client := testClient(t, env, ks)

	for i := 0; i < 10; i++ {
		key := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{byte(i)}, 32)))
		t.Logf("%s", key.ID())
		username := fmt.Sprintf("a%d", i)
		saveUser(t, env, client, key, username, "github")
	}

	resp, err := client.UserSearch(context.TODO(), "", 0)
	require.NoError(t, err)
	require.Equal(t, 10, len(resp.Users))
	require.Equal(t, "a0", resp.Users[0].Name)

	resp, err = client.UserSearch(context.TODO(), "", 1)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Users))
	require.Equal(t, "a0", resp.Users[0].Name)

	resp, err = client.UserSearch(context.TODO(), "a1", 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Users))
	require.Equal(t, "a1", resp.Users[0].Name)

	resp, err = client.UserSearch(context.TODO(), "z", 1)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp.Users))
}

func TestUser(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	env := testEnv(t)
	defer env.closeFn()

	ks := keys.NewMemKeystore()
	client := testClient(t, env, ks)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	saveUser(t, env, client, alice, "alice", "github")

	resp, err := client.User(context.TODO(), alice.ID())
	require.NoError(t, err)
	require.NotNil(t, resp.User)
	require.Equal(t, "alice", resp.User.Name)

	key := keys.GenerateEdX25519Key()
	resp, err = client.User(context.TODO(), key.ID())
	require.NoError(t, err)
	require.Nil(t, resp)
}
