package wormhole_test

import (
	"net/http/httptest"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/users"
)

type env struct {
	clock      tsutil.Clock
	httpServer *httptest.Server
	srv        *server.Server
	ds         dstore.Documents
	users      *users.Users
	client     http.Client
	closeFn    func()
}

func testEnv(t *testing.T) *env {
	clock := tsutil.NewTestClock()
	fi := dstore.NewMem()
	fi.SetClock(clock)
	rds := server.NewRedisTest(clock)
	client := http.NewClient()
	users := testUserStore(t, fi, client, clock)

	srv := server.New(fi, rds, client, clock, server.NewLogger(server.NoLevel))
	srv.SetClock(clock)
	tasks := server.NewTestTasks(srv)
	srv.SetTasks(tasks)
	srv.SetInternalAuth("testtoken")
	srv.SetInternalKey("6a169a699f7683c04d127504a12ace3b326e8b56a61a9b315cf6b42e20d6a44a")
	handler := server.NewHandler(srv)
	httpServer := httptest.NewServer(handler)
	srv.URL = httpServer.URL

	return &env{clock, httpServer, srv, fi, users, client, func() { httpServer.Close() }}
}

func testUserStore(t *testing.T, ds dstore.Documents, client http.Client, clock tsutil.Clock) *users.Users {
	us := users.New(ds, keys.NewSigchains(ds), users.Client(client), users.Clock(clock))
	return us
}
