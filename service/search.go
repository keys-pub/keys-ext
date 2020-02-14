package service

import (
	"context"
)

// UserSearch (RPC) ...
func (s *service) UserSearch(ctx context.Context, req *UserSearchRequest) (*UserSearchResponse, error) {
	users, err := s.searchUser(ctx, req.Query, int(req.Limit), req.Local)
	if err != nil {
		return nil, err
	}

	return &UserSearchResponse{
		Users: users,
	}, nil
}
