package client_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/user"
	"github.com/stretchr/testify/require"
)

func saveUser(t *testing.T, env *env, cl *client.Client, key *keys.EdX25519Key, name string, service string) *keys.Statement {
	url := ""
	api := ""

	id := hex.EncodeToString(sha256.New().Sum([]byte(service + "/" + name))[:8])

	switch service {
	case "github":
		url = fmt.Sprintf("https://gist.github.com/%s/%s", name, id)
		api = "https://api.github.com/gists/" + id
	default:
		t.Fatal("unsupported service in test")
	}

	sc := keys.NewSigchain(key.ID())
	usr, err := user.New(key.ID(), service, name, url, sc.LastSeq()+1)
	require.NoError(t, err)
	st, err := user.NewSigchainStatement(sc, usr, key, env.clock.Now())
	require.NoError(t, err)

	msg, err := usr.Sign(key)
	require.NoError(t, err)

	env.client.SetProxy(api, func(ctx context.Context, req *http.Request, headers []http.Header) http.ProxyResponse {
		return http.ProxyResponse{Body: []byte(githubMock(name, id, msg))}
	})

	err = cl.SigchainSave(context.TODO(), st)
	require.NoError(t, err)

	// err = cl.Check(key)
	// require.NoError(t, err)

	return st
}

func githubMock(name string, id string, msg string) string {
	msg = strings.ReplaceAll(msg, "\n", "")
	return `{
		"id": "` + id + `",
		"files": {
			"gistfile1.txt": {
				"content": "` + msg + `"
			}		  
		},
		"owner": {
			"login": "` + name + `"
		}
	  }`
}

func TestUserSearch(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	env, closeFn := newEnv(t)
	defer closeFn()

	client := newTestClient(t, env)

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

	key1 := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{byte(1)}, 32)))
	resp, err = client.UserSearch(context.TODO(), key1.ID().String(), 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Users))
	require.Equal(t, "a1", resp.Users[0].Name)
}

func TestUser(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	env, closeFn := newEnv(t)
	defer closeFn()

	client := newTestClient(t, env)

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
