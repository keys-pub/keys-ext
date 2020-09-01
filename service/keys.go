package service

import (
	"context"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
)

// Keys (RPC) ...
func (s *service) Keys(ctx context.Context, req *KeysRequest) (*KeysResponse, error) {
	query := strings.TrimSpace(req.Query)

	types := make([]keys.KeyType, 0, len(req.Types))
	for _, t := range req.Types {
		typ, err := keyTypeFromRPC(t)
		if err != nil {
			return nil, err
		}
		types = append(types, typ)
	}

	ks, err := s.vault.Keys(vault.Keys.Types(types...))
	if err != nil {
		return nil, err
	}

	out, err := s.filterKeys(ctx, ks, query)
	if err != nil {
		return nil, err
	}

	return &KeysResponse{
		Keys: out,
	}, nil
}

func containsQuery(query string, key *Key) bool {
	if strings.HasPrefix(key.ID, query) {
		return true
	}
	for _, usr := range key.Users {
		if strings.HasPrefix(usr.ID, query) {
			return true
		}
	}
	return false
}

func (s *service) filterKeys(ctx context.Context, ks []keys.Key, query string) ([]*Key, error) {
	keys := make([]*Key, 0, len(ks))
	for _, k := range ks {
		key, err := s.keyToRPC(ctx, k)
		if err != nil {
			return nil, err
		}
		if query == "" || containsQuery(query, key) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}
