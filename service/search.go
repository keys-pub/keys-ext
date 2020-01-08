package service

import (
	"context"
)

// Search (RPC) ...
func (s *service) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	res, err := s.searchUser(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, 0, len(res))
	for _, res := range res {
		results = append(results, &SearchResult{
			KID:   res.KID.String(),
			Users: userResultsToRPC(res.Users),
		})
	}

	return &SearchResponse{
		Results: results,
	}, nil
}
