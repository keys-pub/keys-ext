package service

import (
	"context"
	"sort"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Keys (RPC) ...
func (s *service) Keys(ctx context.Context, req *KeysRequest) (*KeysResponse, error) {
	sortField := req.SortField
	if sortField == "" {
		sortField = "type"
	}
	sortDirection := req.SortDirection

	kids, err := s.kidsSet(true)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load keystore kids")
	}

	keys, err := s.keys(ctx, kids.IDs(), sortField, sortDirection)
	if err != nil {
		return nil, err
	}

	return &KeysResponse{
		Keys:          keys,
		SortField:     sortField,
		SortDirection: sortDirection,
	}, nil
}

func (s *service) keys(ctx context.Context, kids []keys.ID, sortField string, sortDirection SortDirection) ([]*Key, error) {
	keys := make([]*Key, 0, len(kids))
	for _, kid := range kids {
		key, err := s.key(ctx, kid, true, false)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load keystore key")
		}
		keys = append(keys, key)
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
		if len(pks[i].Users) == 0 && len(pks[j].Users) == 0 {
			return keysSort(pks, "kid", sortDirection, i, j)
		} else if len(pks[i].Users) == 0 {
			return false
		} else if len(pks[j].Users) == 0 {
			return true
		}
		if sortDirection == SortDesc {
			return pks[i].Users[0].Name > pks[j].Users[0].Name
		}
		return pks[i].Users[0].Name <= pks[j].Users[0].Name
	default:
		if sortDirection == SortDesc {
			return pks[i].KID > pks[j].KID
		}
		return pks[i].KID <= pks[j].KID
	}
}
