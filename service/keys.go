package service

import (
	"context"
	"sort"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Keys (RPC) ...
func (s *service) Keys(ctx context.Context, req *KeysRequest) (*KeysResponse, error) {
	sortField := req.SortField
	if sortField == "" {
		sortField = "user"
	}
	sortDirection := req.SortDirection

	types := make([]keys.KeyType, 0, len(req.Types))
	for _, t := range req.Types {
		typ, err := keyTypeFromRPC(t)
		if err != nil {
			return nil, err
		}
		types = append(types, typ)
	}

	ks := s.keyStore()
	out, err := ks.Keys(&keys.Opts{Types: types})
	if err != nil {
		return nil, err
	}

	keys, err := s.keys(ctx, out, req.Query, sortField, sortDirection)
	if err != nil {
		return nil, err
	}

	return &KeysResponse{
		Keys:          keys,
		SortField:     sortField,
		SortDirection: sortDirection,
	}, nil
}

func (s *service) keys(ctx context.Context, ks []keys.Key, query string, sortField string, sortDirection SortDirection) ([]*Key, error) {
	keys := make([]*Key, 0, len(ks))
	for _, k := range ks {
		key, err := s.keyToRPC(ctx, k)
		if err != nil {
			return nil, err
		}
		if query == "" || (key.User != nil && strings.HasPrefix(key.User.ID, query)) || strings.HasPrefix(key.ID, query) {
			keys = append(keys, key)
		}
	}

	switch sortField {
	case "kid", "user", "type":
	default:
		return nil, errors.Errorf("invalid sort field")
	}

	sort.Slice(keys, func(i, j int) bool {
		return keysSort(keys, sortField, sortDirection, i, j)
	})
	return keys, nil
}

func keysSort(pks []*Key, sortField string, sortDirection SortDirection, i, j int) bool {
	switch sortField {
	case "type":
		if pks[i].Type == pks[j].Type {
			return keysSort(pks, "user", sortDirection, i, j)
		}
		if sortDirection == SortDesc {
			return pks[i].Type < pks[j].Type
		}
		return pks[i].Type > pks[j].Type

	case "user":
		if pks[i].User == nil && pks[j].User == nil {
			return keysSort(pks, "kid", sortDirection, i, j)
		} else if pks[i].User == nil {
			return false
		} else if pks[j].User == nil {
			return true
		}
		if sortDirection == SortDesc {
			return pks[i].User.Name > pks[j].User.Name
		}
		return pks[i].User.Name <= pks[j].User.Name
	default:
		if sortDirection == SortDesc {
			return pks[i].ID > pks[j].ID
		}
		return pks[i].ID <= pks[j].ID
	}
}
