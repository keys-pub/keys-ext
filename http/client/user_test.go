package client_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/user"
	"github.com/stretchr/testify/require"
)

func saveUser(t *testing.T, env *env, cl *client.Client, key *keys.EdX25519Key, sc *keys.Sigchain, name string, service string) *keys.Statement {
	st, err := user.MockStatement(key, sc, name, service, env.req, env.clock)
	require.NoError(t, err)
	err = cl.SigchainSave(context.TODO(), st)
	require.NoError(t, err)
	return st
}

func TestUserSearch(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	env := newEnv(t, nil)
	defer env.closeFn()

	client := newTestClient(t, env)

	for i := 0; i < 10; i++ {
		key := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{byte(i)}, 32)))
		sc := keys.NewSigchain(key.ID())
		saveUser(t, env, client, key, sc, fmt.Sprintf("g%d", i), "github")
		saveUser(t, env, client, key, sc, fmt.Sprintf("t%d", i), "twitter")
	}

	resp, err := client.UserSearch(context.TODO(), "", 0)
	require.NoError(t, err)
	require.Equal(t, 20, len(resp.Users))
	require.Equal(t, "g0", resp.Users[0].Name)

	resp, err = client.UserSearch(context.TODO(), "", 1)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Users))
	require.Equal(t, "g0", resp.Users[0].Name)

	resp, err = client.UserSearch(context.TODO(), "g1", 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Users))
	require.Equal(t, "g1", resp.Users[0].Name)

	resp, err = client.UserSearch(context.TODO(), "notfound", 1)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp.Users))

	key1 := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{byte(1)}, 32)))
	resp, err = client.UserSearch(context.TODO(), key1.ID().String(), 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Users))
	require.Equal(t, "g1", resp.Users[0].Name)
	require.Equal(t, "t1", resp.Users[1].Name)

	resp, err = client.UserSearch(context.TODO(), "g1@github", 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Users))
	require.Equal(t, "g1@github", resp.Users[0].ID)

	resp, err = client.UserSearch(context.TODO(), "t1@twitter", 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Users))
	require.Equal(t, "t1@twitter", resp.Users[0].ID)
}

func TestUsers(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	env := newEnv(t, nil)
	defer env.closeFn()

	client := newTestClient(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	sc := keys.NewSigchain(alice.ID())
	saveUser(t, env, client, alice, sc, "alice", "github")

	resp, err := client.Users(context.TODO(), alice.ID())
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Users))
	require.Equal(t, "alice", resp.Users[0].Name)

	key := keys.GenerateEdX25519Key()
	resp, err = client.Users(context.TODO(), key.ID())
	require.NoError(t, err)
	require.Nil(t, resp)
}
