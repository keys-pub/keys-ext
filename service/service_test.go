package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/users"
	"github.com/stretchr/testify/require"
)

func newEnv(t *testing.T, appName string, serverURL string) (*Env, CloseFn) {
	if appName == "" {
		appName = "KeysTest-" + randName()
	}
	env, err := NewEnv(appName, build)
	require.NoError(t, err)
	env.Set(serverCfgKey, serverURL)

	closeFn := func() {
		removeErr := os.RemoveAll(env.AppDir())
		require.NoError(t, removeErr)
	}
	return env, closeFn
}

func randName() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
}

func writeTestFile(t *testing.T) string {
	inPath := keys.RandTempPath() + ".txt"
	writeErr := ioutil.WriteFile(inPath, []byte("test message"), 0644)
	require.NoError(t, writeErr)
	return inPath
}

func testFire(t *testing.T, clock tsutil.Clock) server.Fire {
	fi := dstore.NewMem()
	fi.SetClock(clock)
	return fi
}

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}

type testEnv struct {
	clock  tsutil.Clock
	fi     server.Fire
	client http.Client
	users  *users.Users
}

func newTestEnv(t *testing.T) *testEnv {
	clock := tsutil.NewTestClock()
	fi := testFire(t, clock)
	client := http.NewClient()
	usrs := users.New(fi, keys.NewSigchains(fi), users.Client(client), users.Clock(clock))
	return &testEnv{
		clock:  clock,
		fi:     fi,
		client: client,
		users:  usrs,
	}
}

func newTestService(t *testing.T, tenv *testEnv) (*service, CloseFn) {
	serverEnv := newTestServerEnv(t, tenv)
	appName := "KeysTest-" + randName()

	env, closeFn := newEnv(t, appName, serverEnv.url)
	auth := newAuth(env)

	svc, err := newService(env, Build{Version: "1.2.3", Commit: "deadbeef"}, auth, tenv.client, tenv.clock)
	require.NoError(t, err)

	err = svc.Open()
	require.NoError(t, err)

	closeServiceFn := func() {
		serverEnv.closeFn()
		svc.Close()
		closeFn()
	}

	return svc, closeServiceFn
}

var authPassword = "testpassword"

func testAuthSetup(t *testing.T, service *service) {
	_, err := service.AuthSetup(context.TODO(), &AuthSetupRequest{
		Secret: authPassword,
		Type:   PasswordAuth,
	})
	require.NoError(t, err)
	_, err = service.AuthUnlock(context.TODO(), &AuthUnlockRequest{
		Secret: authPassword,
		Type:   PasswordAuth,
		Client: "test",
	})
	require.NoError(t, err)
}

func testAuthLock(t *testing.T, service *service) {
	_, err := service.AuthLock(context.TODO(), &AuthLockRequest{})
	require.NoError(t, err)
}

func testAuthUnlock(t *testing.T, service *service) {
	_, err := service.AuthUnlock(context.TODO(), &AuthUnlockRequest{
		Secret: authPassword,
		Type:   PasswordAuth,
		Client: "test",
	})
	require.NoError(t, err)
}

// func testAuthVault(t *testing.T, service *service, key *keys.EdX25519Key) {
// 	_, err := service.AuthVault(context.TODO(), &AuthVaultRequest{
// 		Key: encoding.MustEncode(key.Seed()[:], encoding.BIP39),
// 	})
// 	require.NoError(t, err)
// }

func testImportKey(t *testing.T, service *service, key *keys.EdX25519Key) {
	encoded, err := api.EncodeKey(api.NewKey(key), authPassword)
	require.NoError(t, err)
	_, err = service.KeyImport(context.TODO(), &KeyImportRequest{
		In:       []byte(encoded),
		Password: authPassword,
	})
	require.NoError(t, err)
}

func testImportID(t *testing.T, service *service, kid keys.ID) {
	_, err := service.KeyImport(context.TODO(), &KeyImportRequest{
		In: []byte(kid.String()),
	})
	require.NoError(t, err)
}

type testUser struct {
	URL      string
	Response string
}

func userSetupGithub(env *testEnv, service *service, key *keys.EdX25519Key, username string) (*testUser, error) {
	serviceName := "github"
	resp, err := service.UserSign(context.TODO(), &UserSignRequest{
		KID:     key.ID().String(),
		Service: serviceName,
		Name:    username,
	})
	if err != nil {
		return nil, err
	}

	id := hex.EncodeToString(sha256.New().Sum([]byte(serviceName + "/" + username))[:8])
	url := fmt.Sprintf("https://gist.github.com/%s/%s", username, id)
	api := "https://api.github.com/gists/" + id
	body := []byte(githubMock(username, id, resp.Message))
	env.client.SetProxy(api, func(ctx context.Context, req *http.Request) http.ProxyResponse {
		return http.ProxyResponse{Body: body}
	})

	_, err = service.UserAdd(context.TODO(), &UserAddRequest{
		KID:     key.ID().String(),
		Service: serviceName,
		Name:    username,
		URL:     url,
	})
	return &testUser{URL: api, Response: string(body)}, err
}

func testUserSetupGithub(t *testing.T, env *testEnv, service *service, key *keys.EdX25519Key, username string) *testUser {
	tu, err := userSetupGithub(env, service, key, username)
	require.NoError(t, err)
	return tu
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

func redditMock(author string, msg string, subreddit string) string {
	msg = strings.ReplaceAll(msg, "\n", " ")
	return `[{   
		"kind": "Listing",
		"data": {
			"children": [
				{
					"kind": "t3",
					"data": {
						"author": "` + author + `",
						"selftext": "` + msg + `",
						"subreddit": "` + subreddit + `"
					}
				}
			]
		}
    }]`
}

// func mockRedditURL(url string) string {
// 	return url + ".json"
// }

// func testRemoveKey(t *testing.T, service *service, key *keys.EdX25519Key) {
// 	_, err := service.KeyRemove(context.TODO(), &KeyRemoveRequest{
// 		KID: key.ID().String(),
// 	})
// 	require.NoError(t, err)
// }

func testPush(t *testing.T, service *service, key *keys.EdX25519Key) {
	_, err := service.Push(context.TODO(), &PushRequest{
		Key: key.ID().String(),
	})
	require.NoError(t, err)
}

func testPull(t *testing.T, service *service, kid keys.ID) {
	_, err := service.Pull(context.TODO(), &PullRequest{
		Key: kid.String(),
	})
	require.NoError(t, err)
}

// func testUnlock(t *testing.T, service *service) {
// 	_, err := service.AuthUnlock(context.TODO(), &AuthUnlockRequest{
// 		Password: keys.RandPassphrase(12),
// 		Type: PasswordAuth,
//      Client: "test",
// 	})
// 	require.NoError(t, err)
// }

type serverEnv struct {
	url     string
	closeFn func()
}

func newTestServerEnv(t *testing.T, env *testEnv) *serverEnv {
	rds := server.NewRedisTest(env.clock)
	srv := server.New(env.fi, rds, env.client, env.clock, server.NewLogger(server.NoLevel))
	srv.SetClock(env.clock)
	tasks := server.NewTestTasks(srv)
	srv.SetTasks(tasks)
	srv.SetInternalAuth("testtoken")
	_ = srv.SetInternalKey("6a169a699f7683c04d127504a12ace3b326e8b56a61a9b315cf6b42e20d6a44a")
	handler := server.NewHandler(srv)
	testServer := httptest.NewServer(handler)
	srv.URL = testServer.URL

	closeFn := func() {
		testServer.Close()
	}
	return &serverEnv{
		url:     srv.URL,
		closeFn: closeFn,
	}
}

func TestRuntimeStatus(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()

	resp, err := service.RuntimeStatus(context.TODO(), &RuntimeStatusRequest{})
	require.NoError(t, err)
	require.Equal(t, "1.2.3", resp.Version)
}

func TestCheckKeys(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()

	testAuthSetup(t, service)

	testImportKey(t, service, alice)
	testUserSetupGithub(t, env, service, alice, "alice")

	err := service.checkKeys(context.TODO())
	require.NoError(t, err)
}

func TestServiceCheck(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// vault.SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()

	testAuthSetup(t, service)
	require.True(t, service.checking)

	testAuthLock(t, service)
	require.False(t, service.checking)

	testAuthUnlock(t, service)
	require.True(t, service.checking)

	testAuthLock(t, service)
	require.False(t, service.checking)
}
