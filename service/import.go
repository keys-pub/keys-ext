package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
)

// TODO: Difference between pull and import is confusing?

// KeyImport (RPC) imports a key.
func (s *service) KeyImport(ctx context.Context, req *KeyImportRequest) (*KeyImportResponse, error) {
	key, err := keys.ParseKey(req.In, req.Password)
	if err != nil {
		return nil, err
	}
	vk := api.NewKey(key)
	now := s.clock.NowMillis()
	if vk.CreatedAt == 0 {
		vk.CreatedAt = now
	}
	if vk.UpdatedAt == 0 {
		vk.UpdatedAt = now
	}
	out, _, err := s.vault.SaveKey(vk)
	if err != nil {
		return nil, err
	}

	if _, _, err := s.update(ctx, out.ID); err != nil {
		return nil, err
	}

	return &KeyImportResponse{
		KID: out.ID.String(),
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
	vk := api.NewKey(id)
	now := s.clock.NowMillis()
	if vk.CreatedAt == 0 {
		vk.CreatedAt = now
	}
	if vk.UpdatedAt == 0 {
		vk.UpdatedAt = now
	}
	if _, _, err := s.vault.SaveKey(vk); err != nil {
		return err
	}
	if err := s.scs.Index(id); err != nil {
		return err
	}
	return nil
}
