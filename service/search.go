package service

import (
	"context"

	"github.com/pkg/errors"
)

// Search (RPC) ...
func (s *service) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if s.remote == nil {
		return nil, errors.Errorf("no remote set")
	}

	// TODO: Sort
	resp, err := s.remote.Search(req.Query, int(req.Index), int(req.Limit))
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, 0, len(resp.Results))
	for _, res := range resp.Results {
		results = append(results, &SearchResult{
			KID:   res.KID.String(),
			Users: userChecksToRPC(res.Users),
		})
	}

	return &SearchResponse{
		Results: results,
	}, nil
}
