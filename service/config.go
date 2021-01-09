package service

import (
	"context"

	"github.com/keys-pub/keys/dstore"
)

func (s *service) ConfigGet(ctx context.Context, req *ConfigGetRequest) (*ConfigGetResponse, error) {
	path := dstore.Path("config", req.Name)
	doc, err := s.db.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var config Config
	if doc != nil {
		if err := doc.To(&config); err != nil {
			return nil, err
		}
	}

	return &ConfigGetResponse{
		Config: &config,
	}, nil
}

func (s *service) ConfigSet(ctx context.Context, req *ConfigSetRequest) (*ConfigSetResponse, error) {
	// TODO: Validate name
	path := dstore.Path("config", req.Name)
	if err := s.db.Set(ctx, path, dstore.From(req.Config)); err != nil {
		return nil, err
	}
	return &ConfigSetResponse{}, nil
}
