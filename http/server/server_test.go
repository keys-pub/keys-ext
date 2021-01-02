package server_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/user"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	Server  *server.Server
	Handler http.Handler
	// Addr if started
	Addr string
}

// func testFirestore(t *testing.T) Fire {
// 	opts := []option.ClientOption{option.WithCredentialsFile("credentials.json")}
// 	fs, fsErr := firestore.New("firestore://chilltest-3297b", opts...)
// 	require.NoError(t, fsErr)
// 	err := fs.Delete(context.TODO(), "/")
// 	require.NoError(t, err)
// 	return fs
// }

func testFire(t *testing.T, clock tsutil.Clock) server.Fire {
	fi := dstore.NewMem()
	fi.SetClock(clock)
	return fi
}

func TestFireCreatedAt(t *testing.T) {
	clock := tsutil.NewTestClock()
	fi := testFire(t, clock)

	err := fi.Set(context.TODO(), "/test/a", dstore.Data([]byte{0x01}))
	require.NoError(t, err)

	doc, err := fi.Get(context.TODO(), "/test/a")
	require.NoError(t, err)
	require.NotNil(t, doc)

	ftime := doc.CreatedAt.Format(http.TimeFormat)
	require.Equal(t, "Fri, 13 Feb 2009 23:31:30 GMT", ftime)
	ftime = doc.CreatedAt.Format(tsutil.RFC3339Milli)
	require.Equal(t, "2009-02-13T23:31:30.001Z", ftime)
}

type env struct {
	clock    tsutil.Clock
	fi       server.Fire
	client   http.Client
	logLevel server.LogLevel
}

func newEnv(t *testing.T) *env {
	clock := tsutil.NewTestClock()
	fi := testFire(t, clock)
	return newEnvWithFire(t, fi, clock)
}

func newEnvWithFire(t *testing.T, fi server.Fire, clock tsutil.Clock) *env {
	client := http.NewClient()
	return &env{
		clock:    clock,
		fi:       fi,
		client:   client,
		logLevel: server.NoLevel,
	}
}

func newTestServer(t *testing.T, env *env) *testServer {
	rds := server.NewRedisTest(env.clock)
	srv := server.New(env.fi, rds, env.client, env.clock, server.NewLogger(env.logLevel))
	tasks := server.NewTestTasks(srv)
	srv.SetTasks(tasks)
	srv.SetInternalAuth(encoding.MustEncode(keys.RandBytes(32), encoding.Base62))
	srv.SetInternalKey("6a169a699f7683c04d127504a12ace3b326e8b56a61a9b315cf6b42e20d6a44a")
	srv.SetClock(env.clock)
	handler := server.NewHandler(srv)
	return &testServer{
		Server:  srv,
		Handler: handler,
	}
}

func (s *testServer) Serve(req *http.Request) (int, nethttp.Header, string) {
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)
	return rr.Code, rr.Header(), rr.Body.String()
}

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}

type testKeys struct {
	alice    *keys.EdX25519Key
	bob      *keys.EdX25519Key
	channel  *keys.EdX25519Key
	channel2 *keys.EdX25519Key
	frank    *keys.EdX25519Key
}

func testKeysSeeded() testKeys {
	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	bob := keys.NewEdX25519KeyFromSeed(testSeed(0x02))
	channel := keys.NewEdX25519KeyFromSeed(testSeed(0xef))
	channel2 := keys.NewEdX25519KeyFromSeed(testSeed(0xf0))
	frank := keys.NewEdX25519KeyFromSeed(testSeed(0x06))
	return testKeys{
		alice:    alice,
		bob:      bob,
		channel:  channel,
		channel2: channel2,
		frank:    frank,
	}
}

func testKeysRandom() testKeys {
	alice := keys.GenerateEdX25519Key()
	bob := keys.GenerateEdX25519Key()
	channel := keys.GenerateEdX25519Key()
	channel2 := keys.GenerateEdX25519Key()
	frank := keys.GenerateEdX25519Key()
	return testKeys{
		alice:    alice,
		bob:      bob,
		channel:  channel,
		channel2: channel2,
		frank:    frank,
	}
}

func userMock(t *testing.T, key *keys.EdX25519Key, name string, service string, client http.Client, clock tsutil.Clock) *keys.Statement {
	url := ""
	api := ""

	id := hex.EncodeToString(sha256.New().Sum([]byte(service + "/" + name))[:8])

	switch service {
	case "github":
		url = fmt.Sprintf("https://gist.github.com/%s/"+id, name)
		api = "https://api.github.com/gists/" + id
	default:
		t.Fatal("unsupported service in test")
	}

	sc := keys.NewSigchain(key.ID())
	usr, err := user.New(key.ID(), service, name, url, sc.LastSeq()+1)
	require.NoError(t, err)
	st, err := user.NewSigchainStatement(sc, usr, key, clock.Now())
	require.NoError(t, err)

	msg, err := usr.Sign(key)
	require.NoError(t, err)
	client.SetProxy(api, func(ctx context.Context, req *http.Request, headers []http.Header) http.ProxyResponse {
		return http.ProxyResponse{Body: []byte(githubMock(name, "1", msg))}
	})

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

func testJSONMarshal(t *testing.T, i interface{}) []byte {
	b, err := json.Marshal(i)
	require.NoError(t, err)
	return b
}

func testJSONUnmarshal(t *testing.T, b []byte, v interface{}) {
	err := json.Unmarshal(b, v)
	require.NoError(t, err)
}

func TestInternalAuth(t *testing.T) {
	env := newEnv(t)
	srv := newTestServer(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// POST /task/check/:kid
	req, err := http.NewRequest("POST", "/task/check/"+alice.ID().String(), nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"auth failed"}}`, body)

	// Set internal auth token
	srv.Server.SetInternalAuth("testtoken")

	// POST /task/check/:kid (with auth)
	req, err = http.NewRequest("POST", "/task/check/"+alice.ID().String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "testtoken")
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "", body)
}
