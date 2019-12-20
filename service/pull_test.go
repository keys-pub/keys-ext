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
	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService, alice, true)
	testUserSetup(t, env, aliceService, alice.ID(), "alice", true)

	respKeys, err := aliceService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(respKeys.Keys))
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", respKeys.Keys[0].KID)

	// Alice pull (default)
	resp, err := aliceService.Pull(ctx, &PullRequest{})
	require.NoError(t, err)
	require.Equal(t, []string{"a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo"}, resp.KIDs)

	// Bob
	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService, bob, true)
	testUserSetup(t, env, bobService, bob.ID(), "bob", true)

	// Bob pull (all)
	resp, err = bobService.Pull(ctx, &PullRequest{All: true})
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.KIDs))
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", resp.KIDs[0])
	require.Equal(t, "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi", resp.KIDs[1])
	respKeys, err = bobService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(respKeys.Keys))
	require.Equal(t, "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi", respKeys.Keys[0].KID)
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", respKeys.Keys[1].KID)

	// Alice (pull bob KID)
	resp, err = aliceService.Pull(ctx, &PullRequest{KID: "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi"})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.KIDs))
	require.Equal(t, "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi", resp.KIDs[0])
	respKeys, err = aliceService.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(respKeys.Keys))
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", respKeys.Keys[0].KID)
	require.Equal(t, "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi", respKeys.Keys[1].KID)

	// Charlie
	charlieService, charlieCloseFn := newTestService(t, env)
	defer charlieCloseFn()
	testAuthSetup(t, charlieService, charlie, true)
	testUserSetup(t, env, charlieService, charlie.ID(), "charlie", true)

	// Charlie pull (alice)
	resp, err = charlieService.Pull(ctx, &PullRequest{User: "alice@github"})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.KIDs))
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", resp.KIDs[0])
}
