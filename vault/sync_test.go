package vault_test

import (
	"context"
	"testing"
	"time"

	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	db1, closeFn1 := newTestDB(t)
	defer closeFn1()
	db2, closeFn2 := newTestDB(t)
	defer closeFn2()
	testSync(t, db1, db2)
}

func TestSyncMem(t *testing.T) {
	mem1, closeFn1 := newTestMem(t)
	defer closeFn1()
	mem2, closeFn2 := newTestMem(t)
	defer closeFn2()
	testSync(t, mem1, mem2)
}

func testSync(t *testing.T, st1 vault.Store, st2 vault.Store) {
	// vault.SetLogger(vault.NewLogger(vault.DebugLevel))
	var err error
	env := newTestEnv(t, nil) // vault.NewLogger(vault.DebugLevel))
	defer env.closeFn()

	ctx := context.TODO()
	clock := tsutil.NewTestClock()

	// Client #1
	client1 := newTestClient(t, env)

	v1 := vault.New(st1)
	v1.SetClient(client1)

	status, err := v1.SyncStatus()
	require.NoError(t, err)
	require.Nil(t, status)

	key, provision := NewTestVaultKey(t, clock)
	err = v1.Setup(key, provision)
	require.NoError(t, err)
	_, err = v1.Unlock(key)
	require.NoError(t, err)

	err = v1.Set(vault.NewItem("key1", []byte("mysecretdata.1a"), "", time.Now()))
	require.NoError(t, err)

	out, err := v1.Get("key1")
	require.NoError(t, err)
	require.Equal(t, "key1", out.ID)
	require.Equal(t, []byte("mysecretdata.1a"), out.Data)

	// CheckSync not enabled
	synced, err := v1.CheckSync(ctx, time.Duration(0))
	require.NoError(t, err)
	require.False(t, synced)

	err = v1.Sync(ctx)
	require.NoError(t, err)

	status, err = v1.SyncStatus()
	require.NoError(t, err)
	require.NotNil(t, status)

	remote := v1.Remote()

	// Client #2
	client2 := newTestClient(t, env)

	v2 := vault.New(st2)
	v2.SetClient(client2)
	err = v2.Clone(ctx, remote)
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

	history, err := v1.ItemHistory("key1")
	require.NoError(t, err)
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

	// CheckSync
	synced, err = v1.CheckSync(ctx, time.Duration(0))
	require.NoError(t, err)
	require.True(t, synced)

	synced, err = v1.CheckSync(ctx, time.Duration(time.Second))
	require.NoError(t, err)
	require.False(t, synced)

	time.Sleep(time.Millisecond)
	synced, err = v1.CheckSync(ctx, time.Duration(time.Millisecond))
	require.NoError(t, err)
	require.True(t, synced)

	history, err = v1.ItemHistory("key1")
	require.NoError(t, err)
	require.Equal(t, 5, len(history))
	require.Equal(t, []byte("mysecretdata.1a"), history[0].Data)
	require.Equal(t, []byte("mysecretdata.1c"), history[1].Data)
	require.Equal(t, []byte("mysecretdata.1b"), history[2].Data)
	require.Equal(t, []byte("mysecretdata.1d"), history[3].Data)
	require.Nil(t, history[4].Data)

	paths, err := vaultPaths(v1, "/auth")
	require.NoError(t, err)
	expected := []string{
		"/auth/ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums",
	}
	require.Equal(t, expected, paths)

	paths, err = vaultPaths(v1, "/item")
	require.NoError(t, err)
	expected = []string{
		"/item/key1",
		"/item/key2",
	}
	require.Equal(t, expected, paths)

	paths, err = vaultPaths(v1, "/provision")
	require.NoError(t, err)
	expected = []string{
		"/provision/ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums",
	}
	require.Equal(t, expected, paths)

	paths, err = vaultPaths(v1, "/pull")
	require.NoError(t, err)
	expected = []string{
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

	paths, err = vaultPaths(v1, "/sync")
	require.NoError(t, err)
	expected = []string{
		"/sync/lastSync",
		"/sync/pull",
		"/sync/push",
		"/sync/rsalt",
	}
	require.Equal(t, expected, paths)

	cols, err := vault.Collections(v1.Store(), "")
	require.NoError(t, err)
	expected = []string{"/auth", "/item", "/provision", "/pull", "/sync"}
	require.Equal(t, expected, cols)

	cols, err = vault.Collections(v1.Store(), "/pull")
	require.NoError(t, err)
	expected = []string{"/000000000000001", "/000000000000002", "/000000000000003", "/000000000000004", "/000000000000005", "/000000000000006", "/000000000000007", "/000000000000008"}
	require.Equal(t, expected, cols)
}

func TestUnsync(t *testing.T) {
	var err error
	env := newTestEnv(t, nil)
	defer env.closeFn()

	ctx := context.TODO()
	clock := tsutil.NewTestClock()

	db, closeFn := newTestDB(t)
	defer closeFn()

	client := newTestClient(t, env)

	vlt := vault.New(db)
	vlt.SetClient(client)

	key, provision := NewTestVaultKey(t, clock)
	err = vlt.Setup(key, provision)
	require.NoError(t, err)
	_, err = vlt.Unlock(key)
	require.NoError(t, err)

	err = vlt.Set(vault.NewItem("key1", []byte("mysecretdata.1a"), "", time.Now()))
	require.NoError(t, err)

	out, err := vlt.Get("key1")
	require.NoError(t, err)
	require.Equal(t, "key1", out.ID)
	require.Equal(t, []byte("mysecretdata.1a"), out.Data)

	err = vlt.Sync(ctx)
	require.NoError(t, err)

	paths, err := vaultPaths(vlt, dstore.Path("pull"))
	require.NoError(t, err)
	expected := []string{
		"/pull/000000000000001/auth/ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums",
		"/pull/000000000000002/provision/ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums",
		"/pull/000000000000003/item/key1",
	}
	require.Equal(t, expected, paths)

	status, err := vlt.SyncStatus()
	require.NoError(t, err)
	require.NotNil(t, status)
	require.NotEmpty(t, status.KID)
	require.NotEmpty(t, status.SyncedAt)
	rkid := status.KID

	// Add pending
	err = vlt.Set(vault.NewItem("key2", []byte("mysecretdata.2"), "", time.Now()))
	require.NoError(t, err)

	err = vlt.Unsync(ctx)
	require.NoError(t, err)

	items, err := vlt.Items()
	require.NoError(t, err)
	require.Equal(t, 2, len(items))
	require.Equal(t, "key1", items[0].ID)
	require.Equal(t, []byte("mysecretdata.1a"), items[0].Data)
	require.Equal(t, "key2", items[1].ID)
	require.Equal(t, []byte("mysecretdata.2"), items[1].Data)
	out, err = vlt.Get("key1")
	require.NoError(t, err)
	require.Equal(t, "key1", out.ID)
	require.Equal(t, []byte("mysecretdata.1a"), out.Data)

	paths, err = vaultPaths(vlt, dstore.Path("push"))
	require.NoError(t, err)
	expected = []string{
		"/push/000000000000001/auth/ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums",
		"/push/000000000000002/provision/ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums",
		"/push/000000000000003/item/key1",
		"/push/000000000000004/item/key2",
	}
	require.Equal(t, expected, paths)

	status, err = vlt.SyncStatus()
	require.NoError(t, err)
	require.Nil(t, status)

	// Re-sync
	err = vlt.Sync(ctx)
	require.NoError(t, err)

	status, err = vlt.SyncStatus()
	require.NoError(t, err)
	require.NotNil(t, status)
	require.NotEqual(t, rkid, status.KID)

	paths, err = vaultPaths(vlt, dstore.Path("pull"))
	require.NoError(t, err)
	expected = []string{
		"/pull/000000000000001/auth/ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums",
		"/pull/000000000000002/provision/ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums",
		"/pull/000000000000003/item/key1",
		"/pull/000000000000004/item/key2",
	}
	require.Equal(t, expected, paths)

	items, err = vlt.Items()
	require.NoError(t, err)
	require.Equal(t, 2, len(items))
	require.Equal(t, "key1", items[0].ID)
	require.Equal(t, []byte("mysecretdata.1a"), items[0].Data)
	require.Equal(t, "key2", items[1].ID)
	require.Equal(t, []byte("mysecretdata.2"), items[1].Data)
}
