package client

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keysd/http/server"
	"github.com/stretchr/testify/require"
)

const aliceSeed = "win rebuild update term layer transfer gain field prepare unique spider cool present argue grab trend eagle casino peace hockey loop seed desert swear"
const bobSeed = "crane chimney shell unique drink dynamic math pilot letter inflict tattoo curtain primary crystal live return affair husband general cargo chat vintage demand deer"
const groupSeed = "absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter advice comic"

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
	clock   *clock
	client  *Client
	srv     *server.Server
	dst     keys.DocumentStore
	ks      *keys.Keystore
	crypto  keys.CryptoProvider
	uc      *keys.UserContext
	req     *keys.MockRequestor
	closeFn func()
}

func testEnv(t *testing.T) *env {
	clock := newClock()
	fi := keys.NewMem()
	fi.SetTimeNow(clock.Now)
	ns := server.NewMemTestCache(clock.Now)
	req := keys.NewMockRequestor()
	uc := keys.NewTestUserContext(req, clock.Now)

	svr := server.NewServer(fi, ns, uc)
	svr.SetNowFn(clock.Now)
	tasks := server.NewTestTasks(svr)
	svr.SetTasks(tasks)
	svr.SetInternalAuth("testtoken")
	svr.SetAccessFn(func(c server.AccessContext, resource server.AccessResource, action server.AccessAction) server.Access {
		return server.AccessAllow()
	})
	handler := server.NewHandler(svr)
	httpServer := httptest.NewServer(handler)

	// CryptoProvider
	ks := keys.NewMemKeystore()
	ks.SetSigchainStore(keys.NewSigchainStore(keys.NewMem()))
	crypto := saltpack.NewSaltpack(ks)

	client, err := NewClient(httpServer.URL, crypto)
	require.NoError(t, err)
	require.NotNil(t, client.nowFn)
	require.NotNil(t, client.httpClient)
	client.SetHTTPClient(httpServer.Client())
	client.SetTimeNow(clock.Now)
	svr.URL = httpServer.URL
	return &env{clock, client, svr, fi, ks, crypto, uc, req, func() { httpServer.Close() }}
}
