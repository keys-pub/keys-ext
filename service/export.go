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

	if req.Public && typ != SSHExport {
		return nil, errors.Errorf("public key only supported for ssh export")
	}

	key, err := s.vault.Key(id)
	if key == nil {
		return nil, keys.NewErrNotFound(id.String())
	}

	if req.NoPassword && req.Password != "" {
		return nil, errors.Errorf("no password option set with password")
	}

	if !req.NoPassword && !req.Public && req.Password == "" {
		return nil, errors.Errorf("password required for export")
	}

	var out keys.Key
	if req.Public {
		switch key.Type {
		case string(keys.EdX25519):
			sk, err := key.AsEdX25519()
			if err != nil {
				return nil, err
			}
			out = sk.PublicKey()
		case string(keys.X25519):
			bk, err := key.AsX25519()
			if err != nil {
				return nil, err
			}
			out = bk.PublicKey()
		case string(keys.EdX25519Public):
			spk, err := key.AsEdX25519Public()
			if err != nil {
				return nil, err
			}
			out = spk
		case string(keys.X25519Public):
			bpk, err := key.AsX25519Public()
			if err != nil {
				return nil, err
			}
			out = bpk
		default:
			return nil, errors.Errorf("unsupported public option for key export")
		}
	} else {
		switch key.Type {
		case string(keys.EdX25519):
			sk, err := key.AsEdX25519()
			if err != nil {
				return nil, err
			}
			out = sk
		case string(keys.X25519):
			bk, err := key.AsX25519()
			if err != nil {
				return nil, err
			}
			out = bk
		}
	}

	enc, err := exportTypeFromRPC(typ)
	if err != nil {
		return nil, err
	}
	msg, err := keys.EncodeKey(out, enc, req.Password)
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
