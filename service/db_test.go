package service

import (
	"context"
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
	testAuthSetup(t, service, alice, false)
	testRecoverKey(t, service, group, true)
	testPullKey(t, service, group)

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
	require.Equal(t, "/sigchain/a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo-000000000000001", respDocs.Documents[0].Path)
	require.Equal(t, "/sigchain/gqPhYydcdbTzHUdqVrrqBnnAJK9tv3gYbrPKPBynjciM-000000000000001", respDocs.Documents[1].Path)

	respPull, err := service.Documents(ctx, &DocumentsRequest{Path: "/.resource"})
	require.NoError(t, err)
	require.Equal(t, 1, len(respPull.Documents))
	require.Equal(t, "/.resource/sigchain/gqPhYydcdbTzHUdqVrrqBnnAJK9tv3gYbrPKPBynjciM-000000000000001", respPull.Documents[0].Path)
}
