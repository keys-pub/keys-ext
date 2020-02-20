package service

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keysd/http/server"
	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T, serverURL string) (*Config, CloseFn) {
	appName := "KeysTest-" + keys.RandPassphrase(12)
	cfg, err := NewConfig(appName)
	require.NoError(t, err)
	cfg.SetServer(serverURL)

	closeFn := func() {
		removeErr := os.RemoveAll(cfg.AppDir())
		require.NoError(t, removeErr)
	}
	return cfg, closeFn
}

func testFire(t *testing.T, clock *clock) server.Fire {
	fi := keys.NewMem()
	fi.SetTimeNow(clock.Now)
	return fi
}

type testEnv struct {
	clock *clock
	fi    server.Fire
	req   *keys.MockRequestor
	users *keys.UserStore
}

func newTestEnv(t *testing.T) *testEnv {
	clock := newClock()
	fi := testFire(t, clock)
	req := keys.NewMockRequestor()
	users := testUserStore(t, fi, keys.NewSigchainStore(fi), req, clock)
	return &testEnv{
		clock: clock,
		fi:    fi,
		req:   req,
		users: users,
	}
}

func testUserStore(t *testing.T, dst keys.DocumentStore, scs keys.SigchainStore, req *keys.MockRequestor, clock *clock) *keys.UserStore {
	ust, err := keys.NewUserStore(dst, scs, []string{keys.Twitter, keys.Github}, req, clock.Now)
	require.NoError(t, err)
	return ust
}

func newTestService(t *testing.T, env *testEnv) (*service, CloseFn) {
	serverEnv := newTestServerEnv(t, env)

	cfg, closeCfg := testConfig(t, serverEnv.url)
	auth, err := newAuth(cfg)
	require.NoError(t, err)
	svc, err := newService(cfg, Build{Version: "1.2.3", Commit: "deadbeef"}, auth, env.req, env.clock.Now)
	require.NoError(t, err)

	closeFn := func() {
		serverEnv.closeFn()
		svc.Close()
		kr, err := keyring.NewKeyring(cfg.AppName())
		require.NoError(t, err)
		err = kr.Reset()
		require.NoError(t, err)
		closeCfg()
	}

	return svc, closeFn
}

func testAuthSetup(t *testing.T, service *service) {
	password := "testpassword"
	_, err := service.AuthSetup(context.TODO(), &AuthSetupRequest{
		Password: password,
	})
	require.NoError(t, err)
}

func testImportKey(t *testing.T, service *service, key *keys.EdX25519Key) {
	saltpack, err := keys.EncodeKeyToSaltpack(key, "testpassword")
	require.NoError(t, err)
	_, err = service.KeyImport(context.TODO(), &KeyImportRequest{
		In:       []byte(saltpack),
		Password: "testpassword",
	})
	require.NoError(t, err)
}

func testImportID(t *testing.T, service *service, kid keys.ID) {
	_, err := service.KeyImport(context.TODO(), &KeyImportRequest{
		In: []byte(kid.String()),
	})
	require.NoError(t, err)
}

func userSetupGithub(env *testEnv, service *service, key *keys.EdX25519Key, username string) error {
	resp, err := service.UserSign(context.TODO(), &UserSignRequest{
		KID:     key.ID().String(),
		Service: "github",
		Name:    username,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://gist.github.com/%s/1", username)
	env.req.SetResponse(url, []byte(resp.Message))

	_, err = service.UserAdd(context.TODO(), &UserAddRequest{
		KID:     key.ID().String(),
		Service: "github",
		Name:    username,
		URL:     url,
	})
	return err
}

func testUserSetupGithub(t *testing.T, env *testEnv, service *service, key *keys.EdX25519Key, username string) {
	err := userSetupGithub(env, service, key, username)
	require.NoError(t, err)
}

func testRemoveKey(t *testing.T, service *service, key *keys.EdX25519Key) {
	_, err := service.KeyRemove(context.TODO(), &KeyRemoveRequest{
		KID: key.ID().String(),
	})
	require.NoError(t, err)
}

func testPush(t *testing.T, service *service, key *keys.EdX25519Key) {
	_, err := service.Push(context.TODO(), &PushRequest{
		Identity: key.ID().String(),
	})
	require.NoError(t, err)
}

func testPull(t *testing.T, service *service, kid keys.ID) {
	_, err := service.Pull(context.TODO(), &PullRequest{
		Identity: kid.String(),
	})
	require.NoError(t, err)
}

func testUnlock(t *testing.T, service *service) {
	_, err := service.AuthUnlock(context.TODO(), &AuthUnlockRequest{
		Password: keys.RandPassphrase(12),
	})
	require.NoError(t, err)
}

type clock struct {
	t time.Time
}

func newClock() *clock {
	t := keys.TimeFromMillis(1234567890000)
	return &clock{
		t: t,
	}
}

func (c *clock) Now() time.Time {
	c.t = c.t.Add(time.Millisecond)
	return c.t
}

type serverEnv struct {
	url     string
	closeFn func()
}

func newTestServerEnv(t *testing.T, env *testEnv) *serverEnv {
	mc := server.NewMemTestCache(env.clock.Now)
	srv := server.NewServer(env.fi, mc, env.users)
	srv.SetNowFn(env.clock.Now)
	tasks := server.NewTestTasks(srv)
	srv.SetTasks(tasks)
	srv.SetInternalAuth("testtoken")
	srv.SetAccessFn(func(c server.AccessContext, resource server.AccessResource, action server.AccessAction) server.Access {
		return server.AccessAllow()
	})
	handler := server.NewHandler(srv)
	testServer := httptest.NewServer(handler)
	srv.URL = testServer.URL

	closeFn := func() {
		testServer.Close()
	}
	return &serverEnv{
		url:     srv.URL,
		closeFn: closeFn,
	}
}

func spewService(t *testing.T, service *service) {
	iter, iterErr := service.db.Documents(context.TODO(), "", nil)
	require.NoError(t, iterErr)
	spew, err := keys.Spew(iter, nil)
	require.NoError(t, err)
	t.Logf(spew.String())
}

func TestRuntimeStatus(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()

	resp, err := service.RuntimeStatus(context.TODO(), &RuntimeStatusRequest{})
	require.NoError(t, err)
	require.Equal(t, "1.2.3", resp.Version)
}
