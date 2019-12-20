package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service, alice, true)
	testUserSetup(t, env, service, alice.ID(), "alice", true)

	testRecoverKey(t, service, bob, false)
	testUserSetup(t, env, service, bob.ID(), "bob", false)

	testRecoverKey(t, service, charlie, true)
	testRemoveKey(t, service, charlie)

	resp, err := service.Search(ctx, &SearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Results))
	// Alice
	require.Equal(t, alice.ID().String(), resp.Results[0].KID)
	require.Equal(t, 1, len(resp.Results[0].Users))
	require.Equal(t, "alice", resp.Results[0].Users[0].Name)
	// Charlie
	require.Equal(t, charlie.ID().String(), resp.Results[1].KID)
}
