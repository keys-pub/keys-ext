package client_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/users"
	"github.com/stretchr/testify/require"
)

type env struct {
	clock      tsutil.Clock
	fi         server.Fire
	users      *users.Users
	client     http.Client
	logger     server.Logger
	srv        *server.Server
	httpServer *httptest.Server
	handler    http.Handler
}

func newEnv(t *testing.T) (*env, func()) {
	return newEnvWithOptions(t, &envOptions{logLevel: server.NoLevel})
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
	logLevel  server.LogLevel
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
		mem := dstore.NewMem()
		mem.SetClock(opts.clock)
		opts.fi = mem
	}
	rds := server.NewRedisTest(opts.clock)
	client := http.NewClient()
	usrs := users.New(opts.fi, keys.NewSigchains(opts.fi), users.Client(client), users.Clock(opts.clock))

	serverLogger := server.NewLogger(opts.logLevel)
	srv := server.New(opts.fi, rds, client, opts.clock, serverLogger)
	srv.SetClock(opts.clock)
	tasks := server.NewTestTasks(srv)
	srv.SetTasks(tasks)
	srv.SetInternalAuth("testtoken")
	_ = srv.SetInternalKey("6a169a699f7683c04d127504a12ace3b326e8b56a61a9b315cf6b42e20d6a44a")

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
		client:     client,
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

type testKeys struct {
	alice    *keys.EdX25519Key
	bob      *keys.EdX25519Key
	channel  *keys.EdX25519Key
	channel2 *keys.EdX25519Key
}

var alice = keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
var bob = keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
var channel = keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0xef}, 32)))
var channel2 = keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0xf0}, 32)))

func testKeysSeeded() testKeys {
	return testKeys{
		alice:    alice,
		bob:      bob,
		channel:  channel,
		channel2: channel2,
	}
}

func testKeysRandom() testKeys {
	alice := keys.GenerateEdX25519Key()
	bob := keys.GenerateEdX25519Key()
	channel := keys.GenerateEdX25519Key()
	channel2 := keys.GenerateEdX25519Key()
	return testKeys{
		alice:    alice,
		bob:      bob,
		channel:  channel,
		channel2: channel2,
	}
}
