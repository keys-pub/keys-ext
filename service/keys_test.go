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
	service, closeFn := newTestService(t, env)
	defer closeFn()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetupGithub(t, env, service, alice, "alice")
	testPush(t, service, alice)

	testImportKey(t, service, bob)
	testUserSetupGithub(t, env, service, bob, "bob")
	testPush(t, service, bob)

	testImportKey(t, service, charlie)
	testUserSetupGithub(t, env, service, charlie, "charlie")
	testPush(t, service, charlie)

	// Default
	resp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, "user", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 3, len(resp.Keys))
	require.Equal(t, alice.ID().String(), resp.Keys[0].ID)
	require.NotNil(t, resp.Keys[0].User)
	require.Equal(t, "alice", resp.Keys[0].User.Name)
	require.Equal(t, EdX25519, resp.Keys[0].Type)
	require.Equal(t, bob.ID().String(), resp.Keys[1].ID)
	require.NotNil(t, resp.Keys[1].User)
	require.Equal(t, "bob", resp.Keys[1].User.Name)
	require.Equal(t, charlie.ID().String(), resp.Keys[2].ID)
	require.NotNil(t, resp.Keys[2].User)
	require.Equal(t, "charlie", resp.Keys[2].User.Name)

	// KID (asc)
	resp, err = service.Keys(ctx, &KeysRequest{
		SortField: "kid",
	})
	require.NoError(t, err)
	require.Equal(t, "kid", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 3, len(resp.Keys))
	require.Equal(t, alice.ID().String(), resp.Keys[0].ID)
	require.Equal(t, charlie.ID().String(), resp.Keys[1].ID)
	require.Equal(t, bob.ID().String(), resp.Keys[2].ID)

	// KID (desc)
	resp, err = service.Keys(ctx, &KeysRequest{
		SortField:     "kid",
		SortDirection: SortDesc,
	})
	require.NoError(t, err)
	require.Equal(t, "kid", resp.SortField)
	require.Equal(t, SortDesc, resp.SortDirection)
	require.Equal(t, 3, len(resp.Keys))
	require.Equal(t, bob.ID().String(), resp.Keys[0].ID)
	require.Equal(t, charlie.ID().String(), resp.Keys[1].ID)
	require.Equal(t, alice.ID().String(), resp.Keys[2].ID)

	// User (asc)
	resp, err = service.Keys(ctx, &KeysRequest{
		SortField: "user",
	})
	require.NoError(t, err)
	require.Equal(t, "user", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 3, len(resp.Keys))
	require.Equal(t, alice.ID().String(), resp.Keys[0].ID)
	require.Equal(t, bob.ID().String(), resp.Keys[1].ID)
	require.Equal(t, charlie.ID().String(), resp.Keys[2].ID)

	// User (desc)
	resp, err = service.Keys(ctx, &KeysRequest{
		SortField:     "user",
		SortDirection: SortDesc,
	})
	require.NoError(t, err)
	require.Equal(t, "user", resp.SortField)
	require.Equal(t, SortDesc, resp.SortDirection)
	require.Equal(t, 3, len(resp.Keys))
	require.Equal(t, charlie.ID().String(), resp.Keys[0].ID)
	require.Equal(t, bob.ID().String(), resp.Keys[1].ID)
	require.Equal(t, alice.ID().String(), resp.Keys[2].ID)

	// Type
	resp, err = service.Keys(ctx, &KeysRequest{
		SortField: "type",
	})
	require.NoError(t, err)
	require.Equal(t, "type", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 3, len(resp.Keys))
	require.Equal(t, alice.ID().String(), resp.Keys[0].ID)
	require.Equal(t, bob.ID().String(), resp.Keys[1].ID)
	require.Equal(t, charlie.ID().String(), resp.Keys[2].ID)
}

func TestKeysMissingSigchain(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetupGithub(t, env, service, alice, "alice")
	testPush(t, service, alice)

	_, err := service.scs.DeleteSigchain(alice.ID())
	require.NoError(t, err)

	resp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Keys))
}
