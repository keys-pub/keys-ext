package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

// KeyExport (RPC) returns exports a key.
func (s *service) KeyExport(ctx context.Context, req *KeyExportRequest) (*KeyExportResponse, error) {
	id, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}

	typ := req.Type
	if typ == DefaultExportType {
		typ = SaltpackExportType
	}

	if _, err := s.auth.verifyPassword(req.Password); err != nil {
		if err == keyring.ErrInvalidAuth {
			return nil, errors.Errorf("invalid password")
		}
		return nil, err
	}

	switch typ {
	case SaltpackExportType:
		msg, err := s.ks.ExportSaltpack(id, req.Password)
		if err != nil {
			return nil, err
		}
		return &KeyExportResponse{Export: []byte(msg)}, nil
	default:
		return nil, errors.Errorf("unrecognized export type")
	}
}
