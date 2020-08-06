package server_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/request"
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
	fi := docs.NewMem()
	fi.SetClock(clock)
	return fi
}

func TestFireCreatedAt(t *testing.T) {
	clock := tsutil.NewTestClock()
	fi := testFire(t, clock)

	err := fi.Set(context.TODO(), "/test/a", []byte{0x01})
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
	pubSub   server.PubSub
	req      *request.MockRequestor
	logLevel server.LogLevel
}

func newEnv(t *testing.T) *env {
	clock := tsutil.NewTestClock()
	fi := testFire(t, clock)
	return newEnvWithFire(t, fi, clock)
}

func newEnvWithFire(t *testing.T, fi server.Fire, clock tsutil.Clock) *env {
	req := request.NewMockRequestor()
	pubSub := server.NewPubSub()
	return &env{
		clock:    clock,
		fi:       fi,
		req:      req,
		pubSub:   pubSub,
		logLevel: server.ErrLevel,
	}
}

func newTestServer(t *testing.T, env *env) *testServer {
	rds := api.NewRedisTest(env.clock)
	svr := server.New(env.fi, rds, env.req, env.clock, server.NewLogger(env.logLevel))
	tasks := server.NewTestTasks(svr)
	svr.SetTasks(tasks)
	svr.SetInternalAuth(encoding.MustEncode(keys.RandBytes(32), encoding.Base62))
	svr.SetClock(env.clock)
	svr.SetAccessFn(func(c server.AccessContext, resource server.AccessResource, action server.AccessAction) server.Access {
		return server.AccessAllow()
	})
	handler := server.NewHandler(svr)
	return &testServer{
		Server:  svr,
		Handler: handler,
	}
}

func (s *testServer) Serve(req *http.Request) (int, http.Header, string) {
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)
	return rr.Code, rr.Header(), rr.Body.String()
}

func newTestPubSubServer(t *testing.T, env *env) *testPubSubServer {
	pubSub := server.NewPubSub()
	rds := api.NewRedisTest(env.clock)
	svr := server.NewPubSubServer(pubSub, rds, server.NewLogger(server.ErrLevel))
	svr.SetClock(env.clock)
	handler := server.NewPubSubHandler(svr)
	return &testPubSubServer{
		Server:  svr,
		Handler: handler,
	}
}

type testPubSubServer struct {
	Server  *server.PubSubServer
	Handler http.Handler
	// Addr if started
	Addr string
}

func (s *testPubSubServer) Serve(req *http.Request) (int, http.Header, string) {
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)
	return rr.Code, rr.Header(), rr.Body.String()
}

func (s *testPubSubServer) Start() (close func()) {
	server := httptest.NewServer(s.Handler)
	s.Addr = server.Listener.Addr().String()
	return func() {
		server.Close()
	}
}

func (s *testPubSubServer) WebsocketDial(t *testing.T, path string, clock tsutil.Clock, key *keys.EdX25519Key) *websocket.Conn {
	var wsAddr string
	header := http.Header{}

	if key != nil {
		auth, err := api.NewAuth("GET", path, "", clock.Now(), key)
		require.NoError(t, err)
		wsAddr = fmt.Sprintf("ws://%s%s", s.Addr, auth.URL.String())

		header.Set("Authorization", auth.Header())
	} else {
		wsAddr = fmt.Sprintf("ws://%s%s", s.Addr, path)
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsAddr, header)
	require.NoError(t, err)

	return conn
}

func userMock(t *testing.T, key *keys.EdX25519Key, name string, service string, mock *request.MockRequestor, clock tsutil.Clock) *keys.Statement {
	url := ""
	switch service {
	case "github":
		url = fmt.Sprintf("https://gist.github.com/%s/1", name)
	case "twitter":
		url = fmt.Sprintf("https://twitter.com/%s/status/1", name)
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
	mock.SetResponse(url, []byte(msg))

	return st
}

func TestAccess(t *testing.T) {
	env := newEnv(t)
	srv := newTestServer(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	scCount := 0
	srv.Server.SetAccessFn(func(c server.AccessContext, resource server.AccessResource, action server.AccessAction) server.Access {
		switch resource {
		case server.SigchainResource:
			if action == server.Put {
				scCount++
				if scCount == 2 {
					return server.AccessDenyTooManyRequests("sigchain deny test")
				}
			}
		}
		return server.AccessAllow()
	})

	// PUT /sigchain/:kid/:seq (alice, allow)
	aliceSc := keys.NewSigchain(alice.ID())
	aliceSt, err := keys.NewSigchainStatement(aliceSc, []byte("testing"), alice, "", clock.Now())
	require.NoError(t, err)
	err = aliceSc.Add(aliceSt)
	require.NoError(t, err)
	aliceStBytes, err := aliceSt.Bytes()
	require.NoError(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(aliceStBytes))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// PUT /sigchain/:kid/:seq (alice, deny)
	aliceSt2, err := keys.NewSigchainStatement(aliceSc, []byte("testing"), alice, "", clock.Now())
	require.NoError(t, err)
	err = aliceSc.Add(aliceSt2)
	require.NoError(t, err)
	aliceStBytes2, err := aliceSt2.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/2", alice.ID()), bytes.NewReader(aliceStBytes2))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusTooManyRequests, code)
	require.Equal(t, `{"error":{"code":429,"message":"sigchain deny test"}}`, body)

	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	// PUT /:kid/:seq (bob, allow)
	bobSc := keys.NewSigchain(bob.ID())
	bobSt, err := keys.NewSigchainStatement(bobSc, []byte("testing"), bob, "", clock.Now())
	require.NoError(t, err)
	bobAddErr := bobSc.Add(bobSt)
	require.NoError(t, bobAddErr)
	bobStBytes, err := bobSt.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/%s/1", bob.ID()), bytes.NewReader(bobStBytes))
	require.NoError(t, err)
	req.Host = "sigcha.in"
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", body)

	// POST /task/check/:kid
	req, err = http.NewRequest("POST", "/task/check/"+alice.ID().String(), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"no auth token specified"}}`, body)

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
