package service

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keysd/http/client"
	"github.com/keys-pub/keysd/http/server"
	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T) (*Config, CloseFn) {
	appName := "KeysTest-" + keys.RandPassphrase(12)
	cfg, err := NewConfig(appName)
	require.NoError(t, err)

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

func testService(t *testing.T) (*service, CloseFn) {
	clock := newClock()
	fi := testFire(t, clock)
	return testServiceFire(t, fi, clock)
}

func testServiceFire(t *testing.T, fi server.Fire, clock *clock) (*service, CloseFn) {
	cfg, closeCfg := testConfig(t)
	auth, err := newAuth(cfg)
	require.NoError(t, err)
	svc := newService(cfg, Build{Version: "1.2.3", Commit: "deadbeef"}, auth)
	err = svc.Open()
	require.NoError(t, err)
	svc.db.SetTimeNow(clock.Now)

	// Client/server
	testClient := testServerClientFire(t, svc.ks, fi, clock)
	svc.SetRemote(testClient.cl)

	closeFn := func() {
		testClient.closeFn()
		svc.Close()
		kr, krErr := keyring.NewKeyring(cfg.AppName())
		require.NoError(t, krErr)
		reerr := kr.Reset()
		require.NoError(t, reerr)
		closeCfg()
	}

	return svc, closeFn
}

type testKey string

// TODO: Make key IDs start with a,b,c,g,etc
const (
	alice   testKey = "alice"
	bob     testKey = "bob"
	charlie testKey = "charlie"
	group   testKey = "group"
)

func (t testKey) ID() keys.ID {
	switch t {
	case alice:
		return keys.ID("ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg")
	case bob:
		return keys.ID("6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x")
	case charlie:
		return keys.ID("HBtyNnL4mJYQj2QtAb982yokS1Fgy5VYj7Bh5NFBkycS")
	case group:
		return keys.ID("2d8T51ZMqoKsmyKnEAKH1NBtkjCJbjpB2PrUs6SZxsBB")
	default:
		panic("unknown test key")
	}
}

func (t testKey) password() string {
	switch t {
	case alice:
		return "1RH3zBFlk4Yyku"
	case bob:
		return "WuENQNdaJk7cVq"
	case charlie:
		return "yXxV52my3hz3KI"
	case group:
		return "p5HY2Ofy4YPzAa"
	default:
		panic("unknown test key")
	}
}

func (t testKey) pepper() string {
	switch t {
	case alice:
		return "win rebuild update term layer transfer gain field prepare unique spider cool present argue grab trend eagle casino peace hockey loop seed desert swear"
	case bob:
		return "crane chimney shell unique drink dynamic math pilot letter inflict tattoo curtain primary crystal live return affair husband general cargo chat vintage demand deer"
	case charlie:
		return "post hazard hour sad october record orient lesson evolve dizzy jewel conduct diary two argue minute inside circle order floor relief bid attend giraffe"
	case group:
		return "absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter advice comic"
	default:
		panic("unknown test key")
	}
}

func (t testKey) key() keys.Key {
	seed, err := keys.PhraseToBytes(t.seed(), false)
	if err != nil {
		panic(err)
	}
	key, err := keys.NewKey(seed)
	if err != nil {
		panic(err)
	}
	return key
}

func (t testKey) seed() string {
	switch t {
	case alice:
		return "stairs portion summer trade mask nut ostrich hope subway gap daughter sword empty jungle comfort fiscal liberty stadium hint lonely tired found elegant clump"
	case bob:
		return "patient property kitten adapt lunar symptom flag system gun mandate high ice increase disorder party maze earth profit reward lift wool smile test economy"
	case charlie:
		return "motor easy business relax hold now meat rib jungle annual donor spend easy love spirit cable media favorite august vacant tunnel beef small duty"
	case group:
		return "capital club winter remain ladder field enrich tomato predict thought gravity flash ritual since apology person matrix cover grocery turtle hammer desk jungle own"
	default:
		panic("unknown test key")
	}
}

func testAuthSetup(t *testing.T, service *service, tk testKey, publish bool, username string) {
	_, err := service.AuthSetup(context.TODO(), &AuthSetupRequest{
		Password:         tk.password(),
		Pepper:           tk.pepper(),
		PublishPublicKey: publish,
	})
	require.NoError(t, err)
	if username != "" {
		_, userAddErr := service.UserAdd(context.TODO(), &UserAddRequest{
			KID:     tk.ID().String(),
			Service: "test",
			Name:    username,
			URL:     "test://",
			Local:   !publish,
		})
		require.NoError(t, userAddErr)
	}
}

func testRecoverKey(t *testing.T, service *service, tk testKey, publish bool, username string) {
	_, recoverErr := service.KeyRecover(context.TODO(), &KeyRecoverRequest{
		SeedPhrase:       tk.seed(),
		PublishPublicKey: publish,
	})
	require.NoError(t, recoverErr)
	if username != "" {
		_, userAddErr := service.UserAdd(context.TODO(), &UserAddRequest{
			KID:     tk.ID().String(),
			Service: "test",
			Name:    username,
			URL:     "test://",
			Local:   !publish,
		})
		require.NoError(t, userAddErr)
	}
}

func testRemoveKey(t *testing.T, service *service, tk testKey) {
	backupResp, backupErr := service.KeyBackup(context.TODO(), &KeyBackupRequest{
		KID: tk.ID().String(),
	})
	require.NoError(t, backupErr)
	_, removeErr := service.KeyRemove(context.TODO(), &KeyRemoveRequest{
		KID:        tk.ID().String(),
		SeedPhrase: backupResp.SeedPhrase,
	})
	require.NoError(t, removeErr)
}

func testPushKey(t *testing.T, service *service, tk testKey) {
	_, err := service.Push(context.TODO(), &PushRequest{
		KID: tk.ID().String(),
	})
	require.NoError(t, err)
}

func testPullKey(t *testing.T, service *service, tk testKey) {
	_, err := service.Pull(context.TODO(), &PullRequest{
		KID: tk.ID().String(),
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

func testServerClient(t *testing.T, ks *keys.Keystore) (*testClient, server.Fire) {
	clock := newClock()
	fi := testFire(t, clock)
	cl := testServerClientFire(t, ks, fi, clock)
	return cl, fi
}

type testClient struct {
	cl      *client.Client
	closeFn func()
}

func testServerClientFire(t *testing.T, ks *keys.Keystore, fi server.Fire, clock *clock) *testClient {
	mc := server.NewMemTestCache(clock.Now)
	srv := server.NewServer(fi, mc)
	srv.SetNowFn(clock.Now)
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

	crypto := saltpack.NewSaltpack(ks)
	cl, err := client.NewClient(testServer.URL, crypto)
	require.NoError(t, err)
	cl.SetHTTPClient(testServer.Client())
	cl.SetTimeNow(clock.Now)

	return &testClient{
		cl:      cl,
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
	service, closeFn := testService(t)
	defer closeFn()

	resp, err := service.RuntimeStatus(context.TODO(), &RuntimeStatusRequest{})
	require.NoError(t, err)
	require.Equal(t, "1.2.3", resp.Version)
}

func TestFixtures(t *testing.T) {
	require.Equal(t, keys.ID("ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg"), alice.key().ID())
	require.Equal(t, keys.ID("6d35v6U3GfePrTjFwtak5yTUpkEyWA7tQQ2gDzZdX89x"), bob.key().ID())
	require.Equal(t, keys.ID("2d8T51ZMqoKsmyKnEAKH1NBtkjCJbjpB2PrUs6SZxsBB"), group.key().ID())
	require.Equal(t, "stairs portion summer trade mask nut ostrich hope subway gap daughter sword empty jungle comfort fiscal liberty stadium hint lonely tired found elegant clump", alice.seed())
	require.Equal(t, "patient property kitten adapt lunar symptom flag system gun mandate high ice increase disorder party maze earth profit reward lift wool smile test economy", bob.seed())
	require.Equal(t, "capital club winter remain ladder field enrich tomato predict thought gravity flash ritual since apology person matrix cover grocery turtle hammer desk jungle own", group.seed())
}
