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
	if typ == DefaultExport {
		typ = SaltpackExport
	}

	key, err := s.ks.Key(id)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, keys.NewErrNotFound(id.String())
	}

	// TODO: What if we don't have any password auth?
	if req.Password != "" {
		if _, err := s.auth.keyring.UnlockWithPassword(req.Password, false); err != nil {
			if err == keyring.ErrInvalidAuth {
				return nil, errors.Errorf("invalid password")
			}
			return nil, err
		}
	}

	if req.NoPassword && req.Password != "" {
		return nil, errors.Errorf("no-password set with password")
	}

	if !req.NoPassword && !req.Public && req.Password == "" {
		return nil, errors.Errorf("password required for export")
	}

	switch typ {
	case SaltpackExport:
		if req.Public {
			return &KeyExportResponse{Export: []byte(key.ID())}, nil
		}
		return saltpackExportResponse(key, req.Password)
	case SSHExport:
		if key.Type() == keys.EdX25519 {
			if req.Public {
				return sshExportResponseForEdX25519PublicKey(key)
			}
			return sshExportResponseForEdX25519Key(key, req.Password)
		} else if key.Type() == keys.EdX25519Public {
			return sshExportResponseForEdX25519PublicKey(key)
		}

		return nil, errors.Errorf("unsupported key type for ssh export %s", key.Type())

	default:
		return nil, errors.Errorf("unrecognized export type")
	}
}

func saltpackExportResponse(key keys.Key, password string) (*KeyExportResponse, error) {
	msg, err := keys.EncodeKeyToSaltpack(key, password)
	if err != nil {
		return nil, err
	}
	return &KeyExportResponse{Export: []byte(msg)}, nil
}

func sshExportResponseForEdX25519PublicKey(key keys.Key) (*KeyExportResponse, error) {
	pk, err := edX25519PublicKeyFrom(key)
	if err != nil {
		return nil, err
	}
	out := pk.EncodeToSSHAuthorized()
	return &KeyExportResponse{Export: []byte(out)}, nil
}

func sshExportResponseForEdX25519Key(key keys.Key, password string) (*KeyExportResponse, error) {
	k, ok := key.(*keys.EdX25519Key)
	if !ok {
		return nil, errors.Errorf("key type mismatch")
	}
	out, err := k.EncodeToSSH([]byte(password))
	if err != nil {
		return nil, err
	}
	return &KeyExportResponse{Export: []byte(out)}, nil
}

func edX25519PublicKeyFrom(key keys.Key) (*keys.EdX25519PublicKey, error) {
	switch key.Type() {
	case keys.EdX25519:
		k, ok := key.(*keys.EdX25519Key)
		if !ok {
			return nil, errors.Errorf("key type mismatch")
		}
		return k.PublicKey(), nil
	case keys.EdX25519Public:
		pk, ok := key.(*keys.EdX25519PublicKey)
		if !ok {
			return nil, errors.Errorf("key type mismatch")
		}
		return pk, nil
	default:
		return nil, errors.Errorf("invalid key type %s", key.Type())
	}
}
