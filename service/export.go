package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// KeyExport (RPC) exports a key.
func (s *service) KeyExport(ctx context.Context, req *KeyExportRequest) (*KeyExportResponse, error) {
	id, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}

	typ := req.Type
	if typ == DefaultExport {
		typ = SaltpackExport
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
		return nil, errors.Errorf("password required for export")
	}

	if req.Public {
		switch k := key.(type) {
		case *keys.EdX25519Key:
			key = k.PublicKey()
		case *keys.X25519Key:
			key = k.PublicKey()
		case *keys.EdX25519PublicKey, *keys.X25519PublicKey:
			// ok
		default:
			return nil, errors.Errorf("unsupported public option for key export")
		}
	}

	if req.Public && typ != SSHExport {
		return nil, errors.Errorf("public only supported for ssh export")
	}

	enc, err := exportTypeFromRPC(typ)
	if err != nil {
		return nil, err
	}
	msg, err := keys.EncodeKey(key, enc, req.Password)
	if err != nil {
		return nil, err
	}

	return &KeyExportResponse{Export: []byte(msg)}, nil
}

func exportTypeFromRPC(typ ExportType) (keys.Encoding, error) {
	switch typ {
	case SaltpackExport:
		return keys.SaltpackEncoding, nil
	case SSHExport:
		return keys.SSHEncoding, nil
	default:
		return keys.UnknownEncoding, errors.Errorf("unknown export type %s", typ)
	}
}
