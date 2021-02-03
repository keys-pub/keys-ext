package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/keys-pub/keys/dstore"
	"github.com/stretchr/testify/require"
)

func TestDocuments(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	var err error

	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetupGithub(t, env, service, alice, "alice")
	testPush(t, service, alice)

	testImportKey(t, service, bob)
	testUserSetupGithub(t, env, service, bob, "bob")
	testPush(t, service, bob)

	err = service.db.Set(ctx, "/test/key", dstore.Data([]byte("testvalue")))
	require.NoError(t, err)

	respCols, err := service.Collections(ctx, &CollectionsRequest{})
	require.NoError(t, err)

	expectedCols := []*Collection{
		{Path: "/kid"},
		{Path: "/rkl"},
		{Path: "/search"},
		{Path: "/service"},
		{Path: "/sigchain"},
		{Path: "/test"},
		{Path: "/user"},
	}
	require.Equal(t, expectedCols, respCols.Collections)

	respDocs, err := service.Documents(ctx, &DocumentsRequest{Prefix: "/sigchain/"})
	require.NoError(t, err)
	require.Equal(t, 2, len(respDocs.Documents))
	require.Equal(t, fmt.Sprintf("/sigchain/%s-000000000000001", alice.ID()), respDocs.Documents[0].Path)
	require.Equal(t, fmt.Sprintf("/sigchain/%s-000000000000001", bob.ID()), respDocs.Documents[1].Path)
}
