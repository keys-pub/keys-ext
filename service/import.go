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

	if err := s.vault.SaveKey(key); err != nil {
		return nil, err
	}

	if _, _, err := s.update(ctx, key.ID()); err != nil {
		return nil, err
	}

	return &KeyImportResponse{
		KID: key.ID().String(),
	}, nil
}

func (s *service) importID(id keys.ID) error {
	// Check if key already exists and skip if so.
	key, err := s.vault.Key(id)
	if err != nil {
		return err
	}
	if key != nil {
		return nil
	}
	if err := s.vault.SaveKey(id); err != nil {
		return err
	}
	if err := s.scs.Index(id); err != nil {
		return err
	}
	return nil
}
