package service

import (
	"context"
	"encoding/json"

	"github.com/keys-pub/keys/docs"
)

func (s *service) ConfigGet(ctx context.Context, req *ConfigGetRequest) (*ConfigGetResponse, error) {
	path := docs.Path("config", req.Name)
	doc, err := s.db.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return &ConfigGetResponse{}, nil
	}

	var config Config
	if err := json.Unmarshal(doc.Data, &config); err != nil {
		return nil, err
	}

	return &ConfigGetResponse{
		Config: &config,
	}, nil
}

func (s *service) ConfigSet(ctx context.Context, req *ConfigSetRequest) (*ConfigSetResponse, error) {
	// TODO: Validate name
	path := docs.Path("config", req.Name)
	b, err := json.Marshal(req.Config)
	if err != nil {
		return nil, err
	}
	if err := s.db.Set(ctx, path, b); err != nil {
		return nil, err
	}
	return &ConfigSetResponse{}, nil
}
