package vault_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func newTestVaultKey(t *testing.T, clock *tsutil.Clock) (*[32]byte, *vault.Provision) {
	key := keys.Bytes32(bytes.Repeat([]byte{0xFF}, 32))
	id := encoding.MustEncode(bytes.Repeat([]byte{0xFE}, 32), encoding.Base62)
	provision := &vault.Provision{
		ID:        id,
		Type:      vault.UnknownAuth,
		CreatedAt: clock.Now(),
	}
	return key, provision
}

func newTestVault(t *testing.T) *vault.Vault {
	return vault.New(vault.NewMem())
}

func newTestVaultUnlocked(t *testing.T, clock *tsutil.Clock) *vault.Vault {
	vlt := newTestVault(t)
	key, provision := newTestVaultKey(t, clock)
	err := vlt.Setup(key, provision)
	require.NoError(t, err)
	return vlt
}

func newTestVaultDB(t *testing.T) (*vault.DB, func()) {
	path := testPath()
	db := vault.NewDB(path)
	close := func() {
		db.Close()
		_ = os.RemoveAll(path)
	}
	err := db.Open()
	require.NoError(t, err)
	return db, close
}

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}

type testEnv struct {
	clock      *tsutil.Clock
	httpServer *httptest.Server
	srv        *server.Server
	closeFn    func()
}

func newTestEnv(t *testing.T, logger server.Logger) *testEnv {
	if logger == nil {
		logger = client.NewLogger(client.ErrLevel)
	}
	clock := tsutil.NewClock()
	fi := ds.NewMem()
	fi.SetTimeNow(clock.Now)

	rds := server.NewRedisTest(clock.Now)
	srv := server.New(fi, rds, nil, logger)
	srv.SetNowFn(clock.Now)
	srv.SetInternalAuth("testtoken")
	srv.SetAccessFn(func(c server.AccessContext, resource server.AccessResource, action server.AccessAction) server.Access {
		return server.AccessAllow()
	})
	handler := server.NewHandler(srv)
	httpServer := httptest.NewServer(handler)
	srv.URL = httpServer.URL

	return &testEnv{clock, httpServer, srv, func() { httpServer.Close() }}
}

func testPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.vdb", keys.RandFileName()))
}

func testClient(t *testing.T, env *testEnv) *client.Client {
	cl, err := client.New(env.httpServer.URL)
	require.NoError(t, err)
	cl.SetHTTPClient(env.httpServer.Client())
	cl.SetClock(env.clock.Now)
	return cl
}

func TestIsEmpty(t *testing.T) {
	db, closeFn := newTestVaultDB(t)
	defer closeFn()
	vlt := vault.New(db)
	empty, err := vlt.IsEmpty()
	require.NoError(t, err)
	require.True(t, empty)
}

func TestSync(t *testing.T) {
	db1, closeFn1 := newTestVaultDB(t)
	defer closeFn1()
	db2, closeFn2 := newTestVaultDB(t)
	defer closeFn2()
	testSync(t, db1, db2)
}

func TestSyncMem(t *testing.T) {
	testSync(t, vault.NewMem(), vault.NewMem())
}

func testSync(t *testing.T, st1 vault.Store, st2 vault.Store) {
	// vault.SetLogger(vault.NewLogger(vault.DebugLevel))
	var err error
	env := newTestEnv(t, nil) // vault.NewLogger(vault.DebugLevel))
	defer env.closeFn()

	ctx := context.TODO()
	clock := tsutil.NewClock()

	// Client #1
	client1 := testClient(t, env)

	v1 := vault.New(st1)
	v1.SetRemote(client1)

	key, provision := newTestVaultKey(t, clock)
	err = v1.Setup(key, provision)
	require.NoError(t, err)

	err = v1.Set(vault.NewItem("key1", []byte("mysecretdata.1a"), "", time.Now()))
	require.NoError(t, err)

	out, err := v1.Get("key1")
	require.NoError(t, err)
	require.Equal(t, "key1", out.ID)
	require.Equal(t, []byte("mysecretdata.1a"), out.Data)

	err = v1.Sync(ctx)
	require.NoError(t, err)

	// Client #2
	client2 := testClient(t, env)

	v2 := vault.New(st2)
	v2.SetRemote(client2)
	err = v2.InitRemote(ctx, v1.RemoteKey())
	require.NoError(t, err)
	provisionOut, err := v2.Unlock(key)
	require.NoError(t, err)
	require.Equal(t, provision.ID, provisionOut.ID)

	err = v2.Set(vault.NewItem("key2", []byte("mysecretdata.2"), "", time.Now()))
	require.NoError(t, err)

	out, err = v2.Get("key2")
	require.NoError(t, err)
	require.Equal(t, "key2", out.ID)
	require.Equal(t, []byte("mysecretdata.2"), out.Data)

	err = v2.Sync(ctx)
	require.NoError(t, err)

	out, err = v2.Get("key1")
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "key1", out.ID)
	require.Equal(t, []byte("mysecretdata.1a"), out.Data)

	// Update key1 (last sync wins)
	err = v2.Set(vault.NewItem("key1", []byte("mysecretdata.1b"), "", time.Now()))
	require.NoError(t, err)

	err = v1.Set(vault.NewItem("key1", []byte("mysecretdata.1c"), "", time.Now()))
	require.NoError(t, err)

	err = v1.Sync(ctx)
	require.NoError(t, err)

	err = v2.Sync(ctx)
	require.NoError(t, err)

	out, err = v1.Get("key1")
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "key1", out.ID)
	require.Equal(t, []byte("mysecretdata.1c"), out.Data)

	err = v1.Sync(ctx)
	require.NoError(t, err)

	// _ = st1.Spew("", os.Stderr)

	history, err := v1.ItemHistory("key1")
	require.NoError(t, err)
	//vault.SpewItems(versions, os.Stderr)
	require.Equal(t, 3, len(history))
	require.Equal(t, []byte("mysecretdata.1a"), history[0].Data)
	require.Equal(t, []byte("mysecretdata.1c"), history[1].Data)
	require.Equal(t, []byte("mysecretdata.1b"), history[2].Data)

	// Update key1 (without sync)
	err = v1.Set(vault.NewItem("key1", []byte("mysecretdata.1d"), "", time.Now()))
	require.NoError(t, err)

	history, err = v1.ItemHistory("key1")
	require.NoError(t, err)
	require.Equal(t, 4, len(history))
	require.Equal(t, []byte("mysecretdata.1a"), history[0].Data)
	require.Equal(t, []byte("mysecretdata.1c"), history[1].Data)
	require.Equal(t, []byte("mysecretdata.1b"), history[2].Data)
	require.Equal(t, []byte("mysecretdata.1d"), history[3].Data)

	err = v1.Sync(ctx)
	require.NoError(t, err)

	// v2.Spew("", os.Stderr)

	// Delete key1
	del, err := v2.Delete("key1")
	require.NoError(t, err)
	require.True(t, del)

	err = v2.Sync(ctx)
	require.NoError(t, err)

	err = v1.Sync(ctx)
	require.NoError(t, err)

	out, err = v1.Get("key1")
	require.NoError(t, err)
	require.Nil(t, out)

	history, err = v1.ItemHistory("key1")
	require.NoError(t, err)
	require.Equal(t, 5, len(history))
	require.Equal(t, []byte("mysecretdata.1a"), history[0].Data)
	require.Equal(t, []byte("mysecretdata.1c"), history[1].Data)
	require.Equal(t, []byte("mysecretdata.1b"), history[2].Data)
	require.Equal(t, []byte("mysecretdata.1d"), history[3].Data)
	require.Nil(t, history[4].Data)

	paths, err := vault.Paths(v1.Store(), "/pull")
	require.NoError(t, err)
	expected := []string{
		"/pull/000000000000001/auth/ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums",
		"/pull/000000000000002/provision/ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums",
		"/pull/000000000000003/item/key1",
		"/pull/000000000000004/item/key2",
		"/pull/000000000000005/item/key1",
		"/pull/000000000000006/item/key1",
		"/pull/000000000000007/item/key1",
		"/pull/000000000000008/item/key1",
	}
	require.Equal(t, expected, paths)

	paths, err = vault.Paths(v1.Store(), "/item")
	require.NoError(t, err)
	expected = []string{
		"/item/key1",
		"/item/key2",
	}
	require.Equal(t, expected, paths)

	cols, err := vault.Collections(v1.Store(), "")
	require.NoError(t, err)
	require.Equal(t, 5, len(cols))
	require.Equal(t, "/auth", cols[0].Path)
	require.Equal(t, "/db", cols[1].Path)
	require.Equal(t, "/item", cols[2].Path)
	require.Equal(t, "/provision", cols[3].Path)
	require.Equal(t, "/pull", cols[4].Path)
}

func TestErrors(t *testing.T) {
	// vault.SetLogger(vault.NewLogger(vault.DebugLevel))
	var err error
	env := newTestEnv(t, nil) // vault.NewLogger(vault.DebugLevel))
	defer env.closeFn()

	vlt := vault.New(vault.NewMem())

	err = vlt.Set(vault.NewItem("key1", []byte("mysecretdata"), "", time.Now()))
	require.EqualError(t, err, "vault is locked")

	err = vlt.Setup(keys.Rand32(), vault.NewProvision(vault.UnknownAuth))
	require.NoError(t, err)

	err = vlt.Set(vault.NewItem("key1", []byte("mysecretdata"), "", time.Now()))
	require.NoError(t, err)
	vlt.Lock()

	_, err = vlt.Get("key1")
	require.EqualError(t, err, "vault is locked")

	_, err = vlt.Items()
	require.EqualError(t, err, "vault is locked")
	_, err = vlt.ItemHistory("key1")
	require.EqualError(t, err, "vault is locked")
}

func TestUpdate(t *testing.T) {
	db, closeFn := newTestVaultDB(t)
	defer closeFn()
	vlt := vault.New(db)
	testUpdate(t, vlt)
}

func testUpdate(t *testing.T, vlt *vault.Vault) {
	var err error
	key := keys.Rand32()
	provision := vault.NewProvision(vault.UnknownAuth)
	err = vlt.Setup(key, provision)
	require.NoError(t, err)

	items, err := vlt.Items()
	require.NoError(t, err)
	require.Equal(t, 0, len(items))

	out, err := vlt.Get("abc")
	require.NoError(t, err)
	require.Nil(t, out)

	now := time.Now()

	// Set "abc"
	item := vault.NewItem("abc", []byte("password"), "type1", now)
	err = vlt.Set(item)
	require.NoError(t, err)

	out, err = vlt.Get("abc")
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "abc", out.ID)
	require.Equal(t, []byte("password"), out.Data)
	require.Equal(t, tsutil.Millis(now), tsutil.Millis(out.CreatedAt))

	// Update
	item.Data = []byte("newpassword")
	err = vlt.Set(item)
	require.NoError(t, err)

	out, err = vlt.Get("abc")
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "abc", out.ID)
	require.Equal(t, []byte("newpassword"), out.Data)
	require.Equal(t, tsutil.Millis(now), tsutil.Millis(out.CreatedAt))

	// Set "xyz"
	err = vlt.Set(vault.NewItem("xyz", []byte("xpassword"), "type2", time.Now()))
	require.NoError(t, err)

	items, err = vlt.Items()
	require.NoError(t, err)
	require.Equal(t, 2, len(items))
	require.Equal(t, items[0].ID, "abc")
	require.Equal(t, items[1].ID, "xyz")

	// Delete
	ok, err := vlt.Delete("abc")
	require.NoError(t, err)
	require.True(t, ok)

	item3, err := vlt.Get("abc")
	require.NoError(t, err)
	require.Nil(t, item3)

	ok2, err := vlt.Delete("abc")
	require.NoError(t, err)
	require.False(t, ok2)
}

func TestSetupUnlockProvision(t *testing.T) {
	db, closeFn := newTestVaultDB(t)
	defer closeFn()
	vlt := vault.New(db)
	testSetupUnlockProvision(t, vlt)
}

func testSetupUnlockProvision(t *testing.T, vlt *vault.Vault) {
	var err error

	clock := tsutil.NewClock()
	key, provision := newTestVaultKey(t, clock)
	err = vlt.Setup(key, provision)
	require.NoError(t, err)

	err = vlt.Set(vault.NewItem("key1", []byte("password"), "", time.Now()))
	require.NoError(t, err)

	vlt.Lock()

	err = vlt.Set(vault.NewItem("key1", []byte("password"), "", time.Now()))
	require.EqualError(t, err, "vault is locked")

	_, err = vlt.Get("key1")
	require.EqualError(t, err, "vault is locked")

	_, err = vlt.Items()
	require.EqualError(t, err, "vault is locked")

	_, err = vlt.Unlock(key)
	require.NoError(t, err)

	err = vlt.Set(vault.NewItem("key1", []byte("password"), "", time.Now()))
	require.NoError(t, err)

	err = vlt.Lock()
	require.NoError(t, err)

	_, err = vlt.Items()
	require.EqualError(t, err, "vault is locked")

	_, err = vlt.Delete("key1")
	require.EqualError(t, err, "vault is locked")

	key2 := keys.Bytes32(bytes.Repeat([]byte{0x02}, 32))
	_, err = vlt.Unlock(key2)
	require.EqualError(t, err, "invalid vault auth")

	// Unlock
	_, err = vlt.Unlock(key)
	require.NoError(t, err)
	provision2 := vault.NewProvision(vault.UnknownAuth)
	key3 := keys.Rand32()
	err = vlt.Provision(key3, provision2)
	require.NoError(t, err)

	// Deprovision
	ok, err := vlt.Deprovision(provision.ID, false)
	require.NoError(t, err)
	require.True(t, ok)

	paths, err := vault.Paths(vlt.Store(), "/provision")
	require.NoError(t, err)
	require.Equal(t, []string{"/provision/" + provision2.ID}, paths)

	// // Don't deprovision last
	_, err = vlt.Deprovision(provision2.ID, false)
	require.EqualError(t, err, "failed to deprovision: last auth")

	ok, err = vlt.Deprovision(provision2.ID, true)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestSetErrors(t *testing.T) {
	var err error
	clock := tsutil.NewClock()
	vlt := newTestVaultUnlocked(t, clock)

	err = vlt.Set(vault.NewItem("", nil, "", time.Time{}))
	require.EqualError(t, err, "invalid id")
}

func TestProtocol(t *testing.T) {
	st, closeFn := newTestVaultDB(t)
	defer closeFn()
	vlt := vault.New(st)

	var err error

	// Setup
	salt := bytes.Repeat([]byte{0x01}, 32)
	key, err := keys.KeyForPassword("password123", salt)
	require.NoError(t, err)
	provision := &vault.Provision{
		ID:        encoding.MustEncode(bytes.Repeat([]byte{0x02}, 32), encoding.Base62),
		Type:      vault.UnknownAuth,
		CreatedAt: time.Now(),
	}
	err = vlt.Setup(key, provision)
	require.NoError(t, err)

	// Create item
	item := vault.NewItem("testid1", []byte("testpassword"), "", time.Now())
	err = vlt.Set(item)
	require.NoError(t, err)

	paths, err := vault.Paths(st, "")
	require.NoError(t, err)
	require.Equal(t, []string{
		"/auth/0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"/db/increment",
		"/item/testid1",
		"/provision/0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"/push/000000000000001/auth/0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"/push/000000000000002/provision/0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"/push/000000000000003/item/testid1",
	}, paths)

	paths, err = vault.Paths(st, "/auth")
	require.NoError(t, err)
	require.Equal(t, []string{"/auth/" + provision.ID}, paths)

	items, err := vlt.Items()
	require.NoError(t, err)
	require.Equal(t, 1, len(items))
	require.Equal(t, "testid1", items[0].ID)
}
