package wormhole_test

import (
	"net/http/httptest"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys-ext/wormhole"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/request"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/user"
	"github.com/stretchr/testify/require"
)

type env struct {
	clock      tsutil.Clock
	httpServer *httptest.Server
	srv        *server.Server
	ds         docs.Documents
	users      *user.Store
	req        *request.MockRequestor
	closeFn    func()
}

func testEnv(t *testing.T) *env {
	clock := tsutil.NewTestClock()
	fi := docs.NewMem()
	fi.SetClock(clock)
	ns := server.NewRedisTest(clock.Now)
	req := request.NewMockRequestor()
	users := testUserStore(t, fi, req, clock)

	svr := server.New(fi, ns, users, wormhole.NewLogger(wormhole.ErrLevel))
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

func testUserStore(t *testing.T, ds docs.Documents, req request.Requestor, clock tsutil.Clock) *user.Store {
	us, err := user.NewStore(ds, keys.NewSigchainStore(ds), req, clock.Now)
	require.NoError(t, err)
	return us
}
