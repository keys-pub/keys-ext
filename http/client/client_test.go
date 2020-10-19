package client_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/request"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/users"
	"github.com/stretchr/testify/require"
)

type env struct {
	clock      tsutil.Clock
	fi         server.Fire
	users      *users.Users
	req        *request.MockRequestor
	logger     server.Logger
	srv        *server.Server
	httpServer *httptest.Server
	handler    http.Handler
}

func newEnv(t *testing.T) (*env, func()) {
	return newEnvWithOptions(t, nil)
}

type handlerFn func(w http.ResponseWriter, req *http.Request) bool

type proxyHandler struct {
	handlerFn handlerFn
	handler   http.Handler
}

func (p proxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !p.handlerFn(w, req) {
		p.handler.ServeHTTP(w, req)
	}
}

type envOptions struct {
	fi        server.Fire
	clock     tsutil.Clock
	logger    server.Logger
	handlerFn handlerFn
}

func newEnvWithOptions(t *testing.T, opts *envOptions) (*env, func()) {
	if opts == nil {
		opts = &envOptions{}
	}
	if opts.clock == nil {
		opts.clock = tsutil.NewTestClock()
	}
	if opts.fi == nil {
		mem := docs.NewMem()
		mem.SetClock(opts.clock)
		opts.fi = mem
	}
	if opts.logger == nil {
		opts.logger = client.NewLogger(client.ErrLevel)
	}
	rds := server.NewRedisTest(opts.clock)
	req := request.NewMockRequestor()
	usrs := users.New(opts.fi, keys.NewSigchains(opts.fi), users.Requestor(req), users.Clock(opts.clock))

	srv := server.New(opts.fi, rds, req, opts.clock, opts.logger)
	srv.SetClock(opts.clock)
	tasks := server.NewTestTasks(srv)
	srv.SetTasks(tasks)
	srv.SetInternalAuth("testtoken")
	srv.SetAccessFn(func(c server.AccessContext, resource server.AccessResource, action server.AccessAction) server.Access {
		return server.AccessAllow()
	})

	handler := server.NewHandler(srv)
	if opts.handlerFn != nil {
		handler = proxyHandler{
			handlerFn: opts.handlerFn,
			handler:   server.NewHandler(srv),
		}
	}

	httpServer := httptest.NewServer(handler)
	srv.URL = httpServer.URL
	closeFn := func() { httpServer.Close() }

	return &env{
		clock:      opts.clock,
		fi:         opts.fi,
		users:      usrs,
		req:        req,
		logger:     opts.logger,
		srv:        srv,
		httpServer: httpServer,
		handler:    handler,
	}, closeFn
}

func newTestClient(t *testing.T, env *env) *client.Client {
	cl, err := client.New(env.httpServer.URL)
	require.NoError(t, err)
	cl.SetHTTPClient(env.httpServer.Client())
	cl.SetClock(env.clock)
	return cl
}
