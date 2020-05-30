package service

import (
	"context"
	"sort"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

// Item (RPC) returns an item for an ID.
func (s *service) Item(ctx context.Context, req *ItemRequest) (*ItemResponse, error) {
	kr := s.keyring()
	item, err := kr.Get(req.ID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, keys.NewErrNotFound(req.ID)
	}
	return &ItemResponse{
		Item: itemToRPC(item),
	}, nil
}

// Items (RPC) returns list of keyring items.
func (s *service) Items(ctx context.Context, req *ItemsRequest) (*ItemsResponse, error) {
	if req.Query != "" {
		return nil, errors.Errorf("query not implemented")
	}

	kr := s.keyring()
	items, err := kr.List()
	if err != nil {
		return nil, err
	}

	itemsOut := make([]*Item, 0, len(items))
	for _, item := range items {
		itemsOut = append(itemsOut, itemToRPC(item))
	}

	sort.Slice(itemsOut, func(i, j int) bool {
		if itemsOut[i].Type == itemsOut[j].Type {
			return strings.ToLower(itemsOut[i].ID) < strings.ToLower(itemsOut[j].ID)
		}
		return itemsOut[i].Type < itemsOut[j].Type
	})

	return &ItemsResponse{
		Items: itemsOut,
	}, nil
}

func itemToRPC(i *keyring.Item) *Item {
	item := &Item{
		ID:   i.ID,
		Type: i.Type,
	}
	return item
}
