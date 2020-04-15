package service

import "context"

// Search (RPC) ...
func (s *service) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	res, err := s.searchUsersRemote(ctx, req.Query, 0)
	if err != nil {
		return nil, err
	}
	keys := make([]*Key, 0, len(res))
	for _, u := range res {
		kid := u.KID
		typ := keyTypeToRPC(kid.PublicKeyType())
		key := &Key{
			ID:   kid.String(),
			User: apiUserToRPC(u),
			Type: typ,
		}
		keys = append(keys, key)
	}

	return &SearchResponse{
		Keys: keys,
	}, nil
}
