package service

import (
	"context"

	"github.com/keys-pub/keys"
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
	kids := make([]keys.ID, 0, len(resp.Results))
	for _, res := range resp.Results {
		kids = append(kids, res.KID)
	}

	keys, err := s.keys(ctx, kids, "user", SortAsc)
	if err != nil {
		return nil, err
	}

	return &SearchResponse{
		Keys: keys,
	}, nil
}
