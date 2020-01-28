package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// KeyImport (RPC) imports a key.
func (s *service) KeyImport(ctx context.Context, req *KeyImportRequest) (*KeyImportResponse, error) {
	in := req.In
	if utf8.Valid(in) {
		in := strings.TrimSpace(string(in))

		// Try to import key ID
		id, err := keys.ParseID(in)
		if err == nil {
			if err := s.importID(id); err != nil {
				return nil, errors.Wrapf(err, "failed to import key (ID)")
			}
			return &KeyImportResponse{KID: id.String()}, nil
		}

		kid, err := s.importSaltpack(in, req.Password)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to import key")
		}

		return &KeyImportResponse{
			KID: kid.String(),
		}, nil
	}

	return nil, errors.Errorf("unrecognized key format")
}

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

func (s *service) importID(id keys.ID) error {
	// Check if item already exists and skip if so.
	item, err := s.ks.Keyring().Get(id.String())
	if err != nil {
		return err
	}
	if item != nil {
		return nil
	}

	switch id.KeyType() {
	case keys.Ed25519Public:
		spk, err := keys.Ed25519PublicKeyFromID(id)
		if err != nil {
			return err
		}
		return s.ks.SaveSignPublicKey(spk)
	case keys.X25519Public:
		bpk, err := keys.X25519PublicKeyFromID(id)
		if err != nil {
			return err
		}
		return s.ks.SaveBoxPublicKey(bpk)
	default:
		return errors.Errorf("unrecognized key type for %s", id)
	}
}

func (s *service) importSaltpack(in string, password string) (keys.ID, error) {
	key, err := keys.DecodeKeyFromSaltpack(in, password, false)
	if err != nil {
		return "", err
	}
	if err := s.ks.SaveKey(key); err != nil {
		return "", err
	}
	return key.ID(), nil
}
