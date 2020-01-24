package service

import (
	"context"
)

// UserSearch (RPC) ...
func (s *service) UserSearch(ctx context.Context, req *UserSearchRequest) (*UserSearchResponse, error) {
	res, err := s.searchUser(ctx, req.Query, int(req.Limit), req.Local)
	if err != nil {
		return nil, err
	}

	results := make([]*UserSearchResult, 0, len(res))
	for _, r := range res {
		results = append(results, &UserSearchResult{
			KID:   r.KID.String(),
			Users: userResultsToRPC(r.UserResults),
		})
	}

	return &UserSearchResponse{
		Results: results,
	}, nil
}
