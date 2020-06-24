package client_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/request"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/user"
	"github.com/stretchr/testify/require"
)

type env struct {
	clock      *tsutil.Clock
	httpServer *httptest.Server
	srv        *server.Server
	dst        ds.DocumentStore
	users      *user.Store
	req        *request.MockRequestor
	closeFn    func()
}

func testEnv(t *testing.T, logger server.Logger) *env {
	if logger == nil {
		logger = client.NewLogger(client.ErrLevel)
	}
	clock := tsutil.NewClock()
	fi := ds.NewMem()
	fi.SetTimeNow(clock.Now)
	vclock := tsutil.NewClock()
	fi.SetIncrementFn(func(ctx context.Context) (int64, error) {
		return tsutil.Millis(vclock.Now()), nil
	})
	ns := server.NewMemTestCache(clock.Now)
	req := request.NewMockRequestor()
	users := testUserStore(t, fi, req, clock)

	svr := server.New(fi, ns, users, logger)
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

func testClient(t *testing.T, env *env) *client.Client {
	cl, err := client.New(env.httpServer.URL)
	require.NoError(t, err)
	cl.SetHTTPClient(env.httpServer.Client())
	cl.SetClock(env.clock.Now)
	return cl
}

func testUserStore(t *testing.T, ds ds.DocumentStore, req request.Requestor, clock *tsutil.Clock) *user.Store {
	us, err := user.NewStore(ds, keys.NewSigchainStore(ds), req, clock.Now)
	require.NoError(t, err)
	return us
}
