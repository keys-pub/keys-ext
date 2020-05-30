package service

import (
	"context"

	"github.com/keys-pub/keys"
)

// TODO: Difference between pull and import is confusing?

// KeyImport (RPC) imports a key.
func (s *service) KeyImport(ctx context.Context, req *KeyImportRequest) (*KeyImportResponse, error) {
	key, err := keys.ParseKey(req.In, req.Password)
	if err != nil {
		return nil, err
	}

	ks := s.keyStore()
	if err := ks.Save(key); err != nil {
		return nil, err
	}

	// TODO: Should this be optional?
	if _, _, err := s.update(ctx, key.ID()); err != nil {
		return nil, err
	}

	return &KeyImportResponse{
		KID: key.ID().String(),
	}, nil
}

func (s *service) importID(id keys.ID) error {
	// Check if item already exists and skip if so.
	kr := s.keyring()
	ks := keys.NewStore(kr)
	item, err := kr.Get(id.String())
	if err != nil {
		return err
	}
	if item != nil {
		return nil
	}
	return ks.Save(id)
}
