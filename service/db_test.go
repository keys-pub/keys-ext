package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocuments(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice)
	testUserSetup(t, env, service, alice, "alice")
	testPush(t, service, alice)

	testRecoverKey(t, service, bob)
	testUserSetup(t, env, service, bob, "bob")
	testPush(t, service, bob)

	respCols, err := service.Collections(ctx, &CollectionsRequest{})
	require.NoError(t, err)

	expectedCols := []*Collection{
		&Collection{Path: "/.resource"},
		&Collection{Path: "/sigchain"},
	}
	require.Equal(t, expectedCols, respCols.Collections)

	respDocs, err := service.Documents(ctx, &DocumentsRequest{Path: "/sigchain"})
	require.NoError(t, err)

	require.Equal(t, 2, len(respDocs.Documents))
	require.Equal(t, fmt.Sprintf("/sigchain/%s-000000000000001", alice.ID()), respDocs.Documents[0].Path)
	require.Equal(t, fmt.Sprintf("/sigchain/%s-000000000000001", bob.ID()), respDocs.Documents[1].Path)

	respPull, err := service.Documents(ctx, &DocumentsRequest{Path: "/.resource"})
	require.NoError(t, err)
	require.Equal(t, 2, len(respPull.Documents))
	require.Equal(t, fmt.Sprintf("/.resource/sigchain/%s-000000000000001", alice.ID()), respPull.Documents[0].Path)
	require.Equal(t, fmt.Sprintf("/.resource/sigchain/%s-000000000000001", bob.ID()), respPull.Documents[1].Path)
}
