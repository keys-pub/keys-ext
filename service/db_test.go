package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocuments(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))

	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, false, "")
	testRecoverKey(t, service, group, true, "")

	respCols, err := service.Collections(ctx, &CollectionsRequest{})
	require.NoError(t, err)

	expectedCols := []*Collection{
		&Collection{Path: "/sigchain"},
	}
	require.Equal(t, expectedCols, respCols.Collections)

	respDocs, err := service.Documents(ctx, &DocumentsRequest{Path: "/sigchain"})
	require.NoError(t, err)

	require.Equal(t, 2, len(respDocs.Documents))
	require.Equal(t, "/sigchain/2d8T51ZMqoKsmyKnEAKH1NBtkjCJbjpB2PrUs6SZxsBB-000000000000001", respDocs.Documents[0].Path)
	require.Equal(t, "/sigchain/ZoxBoAcN3zUr5A11Uyq1J6pscwKFo2oZSFbwfT7DztXg-000000000000001", respDocs.Documents[1].Path)
}
