package client_test

import (
	"net/http/httptest"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/http/server"
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

func newEnv(t *testing.T, logger server.Logger) *env {
	clock := tsutil.NewTestClock()
	fi := docs.NewMem()
	fi.SetClock(clock)
	return newEnvWithFire(t, fi, clock, logger)
}

func newEnvWithFire(t *testing.T, fi server.Fire, clock tsutil.Clock, logger server.Logger) *env {
	if logger == nil {
		logger = client.NewLogger(client.ErrLevel)
	}
	rds := api.NewRedisTest(clock)
	req := request.NewMockRequestor()
	users := newTestUserStore(t, fi, req, clock)

	srv := server.New(fi, rds, users, logger)
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

func newTestClient(t *testing.T, env *env) *client.Client {
	cl, err := client.New(env.httpServer.URL)
	require.NoError(t, err)
	cl.SetHTTPClient(env.httpServer.Client())
	cl.SetClock(env.clock)
	return cl
}

func newTestUserStore(t *testing.T, ds docs.Documents, req request.Requestor, clock tsutil.Clock) *user.Store {
	us, err := user.NewStore(ds, keys.NewSigchainStore(ds), req, clock)
	require.NoError(t, err)
	return us
}
