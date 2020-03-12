package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// TODO: Difference between pull and import is confusing?

// KeyImport (RPC) imports a key.
func (s *service) KeyImport(ctx context.Context, req *KeyImportRequest) (*KeyImportResponse, error) {
	if !utf8.Valid(req.In) {
		return nil, errors.Errorf("unrecognized key format")
	}

	in := string(req.In)
	in = strings.TrimSpace(in)

	// Try to import key ID
	id, err := keys.ParseID(in)
	if err == nil {
		if err := s.importID(id); err != nil {
			return nil, errors.Wrapf(err, "failed to import key (ID)")
		}
		return &KeyImportResponse{KID: id.String()}, nil
	}

	// TODO: better detection if a ID, so we get a better error message if mistyped
	// TODO: Report password needed on command line, is optional

	logger.Infof("Importing key %s", in)
	kid, err := s.importSaltpack(in, req.Password)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to import key")
	}
	logger.Infof("Imported %s", kid)

	// TODO: Should this be optional?
	if _, _, err := s.update(ctx, kid); err != nil {
		return nil, err
	}

	return &KeyImportResponse{
		KID: kid.String(),
	}, nil

}

func (s *service) importID(id keys.ID) error {
	// Check if item already exists and skip if so.
	kr, err := s.ks.Keyring()
	if err != nil {
		return err
	}
	item, err := kr.Get(id.String())
	if err != nil {
		return err
	}
	if item != nil {
		return nil
	}
	return s.ks.SavePublicKey(id)
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
