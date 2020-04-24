package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPull(t *testing.T) {
	ctx := context.TODO()
	env := newTestEnv(t)

	// Alice
	aliceService, aliceCloseFn := newTestService(t, env, "")
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)
	testUserSetupGithub(t, env, aliceService, alice, "alice")
	testPush(t, aliceService, alice)

	respKeys, err := aliceService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(respKeys.Keys))
	require.Equal(t, alice.ID().String(), respKeys.Keys[0].ID)

	// Bob
	bobService, bobCloseFn := newTestService(t, env, "")
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testUserSetupGithub(t, env, bobService, bob, "bob")
	testPush(t, bobService, bob)

	// Alice (pull bob)
	resp, err := aliceService.Pull(ctx, &PullRequest{Identity: bob.ID().String()})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.KIDs))
	require.Equal(t, bob.ID().String(), resp.KIDs[0])
	respKeys, err = aliceService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(respKeys.Keys))
	require.Equal(t, alice.ID().String(), respKeys.Keys[0].ID)
	require.Equal(t, bob.ID().String(), respKeys.Keys[1].ID)

	// Charlie
	charlieService, charlieCloseFn := newTestService(t, env, "")
	defer charlieCloseFn()
	testAuthSetup(t, charlieService)
	testImportKey(t, charlieService, charlie)
	testUserSetupGithub(t, env, charlieService, charlie, "charlie")
	testPush(t, charlieService, charlie)

	// Charlie (pull alice@github)
	resp, err = charlieService.Pull(ctx, &PullRequest{Identity: "alice@github"})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.KIDs))
	require.Equal(t, alice.ID().String(), resp.KIDs[0])
	respKeys, err = charlieService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(respKeys.Keys))
	require.Equal(t, alice.ID().String(), respKeys.Keys[0].ID)
	require.Equal(t, charlie.ID().String(), respKeys.Keys[1].ID)

	// Alice (pull alice@github)
	resp, err = aliceService.Pull(ctx, &PullRequest{Identity: "alice@github"})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.KIDs))
	require.Equal(t, alice.ID().String(), resp.KIDs[0])
}
