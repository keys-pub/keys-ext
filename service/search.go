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
	resp, err := s.remote.Search(req.Query, int(req.Index), int(req.Limit))
	if err != nil {
		return nil, err
	}

	kids, err := s.kidsSet(false)
	if err != nil {
		return nil, err
	}
	pkids, err := s.scs.KIDs()
	if err != nil {
		return nil, err
	}

	return &SearchResponse{
		Results: searchsToRPC(resp.Results, kids, keys.NewIDSet(pkids...)),
	}, nil
}

func searchToRPC(res *keys.SearchResult, kids *keys.IDSet, pkids *keys.IDSet) *SearchResult {
	if res == nil {
		return nil
	}
	typ := PublicKeyType
	saved := false
	if kids.Contains(res.KID) {
		typ = PrivateKeyType
		saved = true
	} else if pkids.Contains(res.KID) {
		saved = true
	}
	return &SearchResult{
		KID:   res.KID.String(),
		Users: usersToRPC(res.Users),
		Type:  typ,
		Saved: saved,
	}
}

func searchsToRPC(in []*keys.SearchResult, kids *keys.IDSet, pkids *keys.IDSet) []*SearchResult {
	res := make([]*SearchResult, 0, len(in))
	for _, r := range in {
		res = append(res, searchToRPC(r, kids, pkids))
	}
	return res
}
