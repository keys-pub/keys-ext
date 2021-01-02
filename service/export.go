package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
	"github.com/pkg/errors"
)

// KeyExport (RPC) exports a key.
func (s *service) KeyExport(ctx context.Context, req *KeyExportRequest) (*KeyExportResponse, error) {
	id, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}

	key, err := s.vault.Key(id)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, keys.NewErrNotFound(id.String())
	}

	if req.NoPassword && req.Password != "" {
		return nil, errors.Errorf("no password option set with password")
	}

	if !req.NoPassword && !req.Public && req.Password == "" {
		return nil, errors.Errorf("password required for export (or set no password option)")
	}

	switch req.Type {
	case SSHExport:
		var kk keys.Key
		if req.Public {
			kk = key.AsPublic()
		} else {
			kk = key.As()
		}
		out, err := keys.EncodeSSHKey(kk, req.Password)
		if err != nil {
			return nil, err
		}
		return &KeyExportResponse{Export: []byte(out)}, nil
	case DefaultExport:
		if req.Public {
			key = &api.Key{
				ID:     key.ID,
				Public: key.Public,
				Type:   key.Type,
			}
		}
		out, err := api.EncodeKey(key, req.Password)
		if err != nil {
			return nil, err
		}
		return &KeyExportResponse{Export: []byte(out)}, nil
	default:
		return nil, errors.Errorf("unknown export type %s", req.Type)
	}
}
