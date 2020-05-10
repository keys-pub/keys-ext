package client_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/user"
	"github.com/keys-pub/keys/util"
	"github.com/keys-pub/keysd/http/client"
	"github.com/keys-pub/keysd/http/server"
	"github.com/stretchr/testify/require"
)

type clock struct {
	t time.Time
}

func newClock() *clock {
	return newClockAt(1234567890000)
}

func newClockAt(ts int64) *clock {
	t := util.TimeFromMillis(ts)
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
	dst        ds.DocumentStore
	users      *user.Store
	req        *util.MockRequestor
	closeFn    func()
}

func testEnv(t *testing.T, logger server.Logger) *env {
	if logger == nil {
		logger = client.NewLogger(client.ErrLevel)
	}
	clock := newClock()
	fi := ds.NewMem()
	fi.SetTimeNow(clock.Now)
	ns := server.NewMemTestCache(clock.Now)
	req := util.NewMockRequestor()
	users := testUserStore(t, fi, req, clock)

	svr := server.NewServer(fi, ns, users, logger)
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

func testClient(t *testing.T, env *env, ks *keys.Store) *client.Client {
	cl, err := client.NewClient(env.httpServer.URL, ks)
	require.NoError(t, err)
	cl.SetHTTPClient(env.httpServer.Client())
	cl.SetTimeNow(env.clock.Now)
	return cl
}

func testUserStore(t *testing.T, ds ds.DocumentStore, req util.Requestor, clock *clock) *user.Store {
	us, err := user.NewStore(ds, keys.NewSigchainStore(ds), req, clock.Now)
	require.NoError(t, err)
	return us
}
