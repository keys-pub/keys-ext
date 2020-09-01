package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetup(t, env, service, alice, "alice", "github")
	testPush(t, service, alice)

	testImportKey(t, service, bob)
	testUserSetup(t, env, service, bob, "bob", "github")
	testPush(t, service, bob)

	resp, err := service.Search(ctx, &SearchRequest{Query: "alice"})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Keys))
	require.Equal(t, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077", resp.Keys[0].ID)
}
