package wormhole_test

import (
	"net/http/httptest"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys-ext/wormhole"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/request"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/users"
)

type env struct {
	clock      tsutil.Clock
	httpServer *httptest.Server
	srv        *server.Server
	ds         dstore.Documents
	users      *users.Users
	req        *request.MockRequestor
	closeFn    func()
}

func testEnv(t *testing.T) *env {
	clock := tsutil.NewTestClock()
	fi := dstore.NewMem()
	fi.SetClock(clock)
	rds := server.NewRedisTest(clock)
	req := request.NewMockRequestor()
	users := testUserStore(t, fi, req, clock)

	srv := server.New(fi, rds, req, clock, wormhole.NewLogger(wormhole.ErrLevel))
	srv.SetClock(clock)
	tasks := server.NewTestTasks(srv)
	srv.SetTasks(tasks)
	srv.SetInternalAuth("testtoken")
	srv.SetAccessFn(func(c server.AccessContext, resource server.AccessResource, action server.AccessAction) server.Access {
		return server.AccessAllow()
	})
	handler := server.NewHandler(srv)
	httpServer := httptest.NewServer(handler)
	srv.URL = httpServer.URL

	return &env{clock, httpServer, srv, fi, users, req, func() { httpServer.Close() }}
}

func testUserStore(t *testing.T, ds dstore.Documents, req request.Requestor, clock tsutil.Clock) *users.Users {
	us := users.New(ds, keys.NewSigchains(ds), users.Requestor(req), users.Clock(clock))
	return us
}
