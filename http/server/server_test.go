package server_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/keys-pub/keysd/http/server"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

type clock struct {
	t    time.Time
	tick time.Duration
}

func newClock() *clock {
	return newClockAt(1234567890000)
}

func (c *clock) setTick(tick time.Duration) {
	c.tick = tick
}

func newClockAt(ts keys.TimeMs) *clock {
	t := keys.TimeFromMillis(ts)
	return &clock{
		t:    t,
		tick: time.Millisecond,
	}
}

// func newClockAtNow() *clock {
// 	return &clock{
// 		t:    time.Now(),
// 		tick: time.Millisecond,
// 	}
// }

func (c *clock) Now() time.Time {
	c.t = c.t.Add(c.tick)
	return c.t
}

type testServer struct {
	Server  *server.Server
	Handler http.Handler
	// Addr if started
	Addr string
}

// func testFirestore(t *testing.T) Fire {
// 	opts := []option.ClientOption{option.WithCredentialsFile("credentials.json")}
// 	fs, fsErr := firestore.NewFirestore("firestore://chilltest-3297b", opts...)
// 	require.NoError(t, fsErr)
// 	err := fs.Delete(context.TODO(), "/")
// 	require.NoError(t, err)
// 	return fs
// }

func testFire(t *testing.T, clock *clock) server.Fire {
	fi := keys.NewMem()
	fi.SetTimeNow(clock.Now)
	return fi
}

func TestFireCreatedAt(t *testing.T) {
	clock := newClock()
	fi := testFire(t, clock)

	err := fi.Set(context.TODO(), "/test/a", []byte{0x01})
	require.NoError(t, err)

	doc, err := fi.Get(context.TODO(), "/test/a")
	require.NoError(t, err)
	require.NotNil(t, doc)

	ftime := doc.CreatedAt.Format(http.TimeFormat)
	require.Equal(t, "Fri, 13 Feb 2009 23:31:30 GMT", ftime)
	ftime = doc.CreatedAt.Format(keys.RFC3339Milli)
	require.Equal(t, "2009-02-13T23:31:30.001Z", ftime)
}

func testUserStore(t *testing.T, ds keys.DocumentStore, req keys.Requestor, clock *clock) *keys.UserStore {
	us, err := keys.NewUserStore(ds, keys.NewSigchainStore(ds), req, clock.Now)
	require.NoError(t, err)
	return us
}

type env struct {
	clock  *clock
	fi     server.Fire
	pubSub server.PubSub
	users  *keys.UserStore
	req    *keys.MockRequestor
}

func newEnv(t *testing.T) *env {
	clock := newClock()
	fi := testFire(t, clock)
	req := keys.NewMockRequestor()
	pubSub := server.NewPubSub()
	users := testUserStore(t, fi, req, clock)
	return &env{
		clock:  clock,
		fi:     fi,
		req:    req,
		pubSub: pubSub,
		users:  users,
	}
}

func newTestServer(t *testing.T, env *env) *testServer {
	mc := server.NewMemTestCache(env.clock.Now)
	svr := server.NewServer(env.fi, mc, env.pubSub, env.users)
	tasks := server.NewTestTasks(svr)
	svr.SetTasks(tasks)
	svr.SetInternalAuth(keys.RandIDString())
	svr.SetNowFn(env.clock.Now)
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

func (s *testServer) Start() (close func()) {
	server := httptest.NewServer(s.Handler)
	s.Addr = server.Listener.Addr().String()
	return func() {
		server.Close()
	}
}

func (s *testServer) WebsocketDial(t *testing.T, path string, clock *clock, key *keys.EdX25519Key) *websocket.Conn {
	t.Logf("Server addr: %s", s.Addr)
	client, err := net.Dial("tcp", s.Addr)
	require.NoError(t, err)

	var auth *api.Auth
	if key != nil {
		a, err := api.NewAuth("GET", path, clock.Now(), key)
		require.NoError(t, err)
		auth = a
	}

	wsAddr := fmt.Sprintf("ws://%s%s", s.Addr, auth.URL.String())
	t.Logf("WS addr: %s", wsAddr)
	config, err := websocket.NewConfig(wsAddr, "http://localhost")
	require.NoError(t, err)
	if auth != nil {
		config.Header.Set("Authorization", auth.Header())
	}

	conn, err := websocket.NewClient(config, client)
	require.NoError(t, err)
	return conn
}

func userMock(t *testing.T, users *keys.UserStore, key *keys.EdX25519Key, name string, service string, mock *keys.MockRequestor) *keys.Statement {
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
	usr, err := keys.NewUser(users, key.ID(), service, name, url, sc.LastSeq()+1)
	require.NoError(t, err)
	st, err := keys.NewUserSigchainStatement(sc, usr, key, users.Now())
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
	aliceStBytes := aliceSt.Bytes()
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
	aliceStBytes2 := aliceSt2.Bytes()
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
	bobStBytes := bobSt.Bytes()
	req, err = http.NewRequest("PUT", fmt.Sprintf("/%s/1", bob.ID()), bytes.NewReader(bobStBytes))
	require.NoError(t, err)
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
