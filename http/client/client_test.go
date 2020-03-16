package client

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/server"
	"github.com/stretchr/testify/require"
)

type clock struct {
	t time.Time
}

func newClock() *clock {
	return newClockAt(1234567890000)
}

func newClockAt(ts keys.TimeMs) *clock {
	t := keys.TimeFromMillis(ts)
	return &clock{
		t: t,
	}
}

func (c *clock) Now() time.Time {
	c.t = c.t.Add(time.Millisecond)
	return c.t
}

type env struct {
	clock      *clock
	httpServer *httptest.Server
	srv        *server.Server
	dst        keys.DocumentStore
	users      *keys.UserStore
	req        *keys.MockRequestor
	closeFn    func()
}

func testEnv(t *testing.T) *env {
	clock := newClock()
	fi := keys.NewMem()
	fi.SetTimeNow(clock.Now)
	ns := server.NewMemTestCache(clock.Now)
	req := keys.NewMockRequestor()
	users := testUserStore(t, fi, req, clock)

	svr := server.NewServer(fi, ns, users)
	svr.SetNowFn(clock.Now)
	tasks := server.NewTestTasks(svr)
	svr.SetTasks(tasks)
	svr.SetInternalAuth("testtoken")
	svr.SetAccessFn(func(c server.AccessContext, resource server.AccessResource, action server.AccessAction) server.Access {
		return server.AccessAllow()
	})
	handler := server.NewHandler(svr)
	httpServer := httptest.NewServer(handler)
	svr.URL = httpServer.URL

	return &env{clock, httpServer, svr, fi, users, req, func() { httpServer.Close() }}
}

func testClient(t *testing.T, env *env, ks *keys.Keystore) *Client {
	client, err := NewClient(env.httpServer.URL, ks)
	require.NoError(t, err)
	require.NotNil(t, client.nowFn)
	require.NotNil(t, client.httpClient)
	client.SetHTTPClient(env.httpServer.Client())
	client.SetTimeNow(env.clock.Now)
	return client
}

func testUserStore(t *testing.T, ds keys.DocumentStore, req keys.Requestor, clock *clock) *keys.UserStore {
	us, err := keys.NewUserStore(ds, keys.NewSigchainStore(ds), req, clock.Now)
	require.NoError(t, err)
	return us
}
