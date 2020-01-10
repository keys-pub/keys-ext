package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// KeyExport (RPC) returns exports a key.
func (s *service) KeyExport(ctx context.Context, req *KeyExportRequest) (*KeyExportResponse, error) {
	key, err := s.parseKey(req.KID)
	if err != nil {
		return nil, err
	}
	switch req.Type {
	case SaltpackPwExportType, DefaultExportType:
		keyBackup := seedToSaltpack(req.Password, key.Seed()[:])
		return &KeyExportResponse{
			Export: []byte(keyBackup),
		}, nil
	default:
		return nil, errors.Errorf("unrecognized export type %q", req.Type)
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

	hrp, _, err := id.Decode()
	if err != nil {
		return err
	}
	switch hrp {
	case keys.SignKeyType:
		spk, err := keys.SignPublicKeyFromID(id)
		if err != nil {
			return err
		}
		return s.ks.SaveSignPublicKey(spk)
	default:
		return errors.Errorf("unrecognized key id type %s", hrp)
	}
}

func seedToSaltpack(password string, seed []byte) string {
	out := keys.EncryptWithPassword(seed[:], password)
	return keys.EncodeSaltpackMessage(out, "")
}

func saltpackToSeed(password string, msg string) ([]byte, error) {
	b, err := keys.DecodeSaltpackMessage(msg, "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse saltpack")
	}
	seed, err := keys.DecryptWithPassword(b, password)
	if err != nil {
		return nil, err
	}
	return seed, nil
}

func (s *service) importSaltpack(in string, password string) (keys.ID, error) {
	seed, err := saltpackToSeed(password, in)
	if err != nil {
		return "", err
	}
	if len(seed) != 32 {
		return "", errors.Errorf("invalid sign key seed bytes in saltpack message")
	}

	key, err := keys.NewSignKeyFromSeed(keys.Bytes32(seed))
	if err != nil {
		return "", err
	}

	existing, err := s.ks.SignKey(key.ID())
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", errors.Errorf("key already exists")
	}

	if err := s.ks.SaveSignKey(key); err != nil {
		return "", err
	}
	return key.ID(), nil
}

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
