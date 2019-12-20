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
	uc    *keys.UserContext
}

func newTestEnv(t *testing.T) *testEnv {
	clock := newClock()
	fi := testFire(t, clock)
	req := keys.NewMockRequestor()
	uc := keys.NewTestUserContext(req, clock.Now)
	return &testEnv{
		clock: clock,
		fi:    fi,
		req:   req,
		uc:    uc,
	}
}

func newTestService(t *testing.T, env *testEnv) (*service, CloseFn) {
	serverEnv := newTestServerEnv(t, env)

	cfg, closeCfg := testConfig(t, serverEnv.url)
	auth, err := newAuth(cfg)
	require.NoError(t, err)
	svc, err := newService(cfg, Build{Version: "1.2.3", Commit: "deadbeef"}, auth, env.uc, env.clock.Now)
	require.NoError(t, err)

	closeFn := func() {
		serverEnv.closeFn()
		svc.Close()
		kr, krErr := keyring.NewKeyring(cfg.AppName())
		require.NoError(t, krErr)
		reerr := kr.Reset()
		require.NoError(t, reerr)
		closeCfg()
	}

	return svc, closeFn
}

func testAuthSetup(t *testing.T, service *service, key keys.Key, publish bool) {
	password := testPasswordForKey(key)
	recovery := testBackupForKey(key)

	_, err := service.AuthSetup(context.TODO(), &AuthSetupRequest{
		Password:  password,
		KeyBackup: recovery,
	})
	require.NoError(t, err)
	if publish {
		_, err := service.Push(context.TODO(), &PushRequest{
			KID: key.ID().String(),
		})
		require.NoError(t, err)
	}
}

func testRecoverKey(t *testing.T, service *service, key keys.Key, publish bool) {
	_, err := service.KeyRecover(context.TODO(), &KeyRecoverRequest{
		SeedPhrase:       keys.SeedPhrase(key),
		PublishPublicKey: publish,
	})
	require.NoError(t, err)
}

func testUserSetup(t *testing.T, env *testEnv, service *service, kid keys.ID, username string, publish bool) {
	resp, err := service.UserSign(context.TODO(), &UserSignRequest{
		KID:     kid.String(),
		Service: "github",
		Name:    username,
	})
	require.NoError(t, err)

	url := fmt.Sprintf("https://gist.github.com/%s/1", username)
	env.req.SetResponse(url, []byte(resp.Message))

	_, err = service.UserAdd(context.TODO(), &UserAddRequest{
		KID:     kid.String(),
		Service: "github",
		Name:    username,
		URL:     url,
		Local:   !publish,
	})
	require.NoError(t, err)
}

func testRemoveKey(t *testing.T, service *service, key keys.Key) {
	backupResp, err := service.KeyBackup(context.TODO(), &KeyBackupRequest{
		KID: key.ID().String(),
	})
	require.NoError(t, err)
	_, err = service.KeyRemove(context.TODO(), &KeyRemoveRequest{
		KID:        key.ID().String(),
		SeedPhrase: backupResp.SeedPhrase,
	})
	require.NoError(t, err)
}

func testPushKey(t *testing.T, service *service, key keys.Key) {
	_, err := service.Push(context.TODO(), &PushRequest{
		KID: key.ID().String(),
	})
	require.NoError(t, err)
}

func testPullKey(t *testing.T, service *service, key keys.Key) {
	_, err := service.Pull(context.TODO(), &PullRequest{
		KID: key.ID().String(),
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
	srv := server.NewServer(env.fi, mc, env.uc)
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
