package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeys(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	ctx := context.TODO()

	// Alice
	service, closeFn := newTestService(t, env, "")
	defer closeFn()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetup(t, env, service, alice, "alice", "github")
	testPush(t, service, alice)

	testImportKey(t, service, bob)
	testUserSetup(t, env, service, bob, "bob", "github")
	testPush(t, service, bob)

	testImportKey(t, service, charlie)
	testUserSetup(t, env, service, charlie, "charlie", "github")
	testPush(t, service, charlie)

	// Default
	resp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 3, len(resp.Keys))
	// Alice
	require.Equal(t, alice.ID().String(), resp.Keys[0].ID)
	require.Equal(t, 1, len(resp.Keys[0].Users))
	require.Equal(t, "alice", resp.Keys[0].Users[0].Name)
	require.Equal(t, EdX25519, resp.Keys[0].Type)
	// Charlie
	require.Equal(t, charlie.ID().String(), resp.Keys[1].ID)
	require.Equal(t, 1, len(resp.Keys[1].Users))
	require.Equal(t, "charlie", resp.Keys[1].Users[0].Name)
	// Bob
	require.Equal(t, bob.ID().String(), resp.Keys[2].ID)
	require.Equal(t, 1, len(resp.Keys[2].Users))
	require.Equal(t, "bob", resp.Keys[2].Users[0].Name)

}

func TestKeysMissingSigchain(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetup(t, env, service, alice, "alice", "github")
	testPush(t, service, alice)

	_, err := service.scs.Delete(alice.ID())
	require.NoError(t, err)

	resp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Keys))
}
