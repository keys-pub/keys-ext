package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Key (RPC) ...
func (s *service) Key(ctx context.Context, req *KeyRequest) (*KeyResponse, error) {
	var kid keys.ID
	if req.User != "" {
		usr, err := s.searchUserExact(ctx, req.User, true)
		if err != nil {
			return nil, err
		}
		if usr == nil {
			return &KeyResponse{}, nil
		}
		kid = usr.User.KID
	} else {
		k, err := s.parseKID(req.KID)
		if err != nil {
			return nil, err
		}
		kid = k
	}

	key, err := s.loadKey(ctx, kid)
	if err != nil {
		return nil, err
	}

	return &KeyResponse{
		Key: key,
	}, nil
}

// Emoji for KeyType.
func Emoji(key keys.Key) string {
	switch key.Type() {
	case keys.Ed25519:
		return "🖋️"
	case keys.Ed25519Public:
		return "🖋️"
	case keys.Curve25519:
		return "🔑"
	case keys.Curve25519Public:
		return "🔑"
	default:
		return "❓"
	}
}

func (s *service) loadKey(ctx context.Context, id keys.ID) (*Key, error) {
	key, err := s.ks.Key(id)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, nil
	}
	return s.keyToRPC(ctx, key, true)
}

var keyTypeStrings = []string{
	string(keys.Ed25519),
	string(keys.Ed25519Public),
	string(keys.Curve25519),
	string(keys.Curve25519Public),
}

func parseKeyType(s string) (KeyType, error) {
	switch s {
	case string(keys.Ed25519):
		return Ed25519, nil
	case string(keys.Ed25519Public):
		return Ed25519Public, nil
	case string(keys.Curve25519):
		return Curve25519, nil
	case string(keys.Curve25519Public):
		return Curve25519Public, nil
	default:
		return UnknownKeyType, errors.Errorf("unsupported key type %s", s)
	}
}

func keyTypeFromRPC(t KeyType) (keys.KeyType, error) {
	switch t {
	case Ed25519:
		return keys.Ed25519, nil
	case Ed25519Public:
		return keys.Ed25519Public, nil
	case Curve25519:
		return keys.Curve25519, nil
	case Curve25519Public:
		return keys.Curve25519Public, nil
	default:
		return "", errors.Errorf("unsupported key type")
	}
}

func keyTypeToRPC(t keys.KeyType) KeyType {
	switch t {
	case keys.Ed25519:
		return Ed25519
	case keys.Ed25519Public:
		return Ed25519Public
	case keys.Curve25519:
		return Curve25519
	case keys.Curve25519Public:
		return Curve25519Public
	default:
		return UnknownKeyType
	}
}

func (s *service) keyToRPC(ctx context.Context, key keys.Key, saved bool) (*Key, error) {
	if key == nil {
		return nil, nil
	}
	result, err := s.users.Get(ctx, key.ID())
	if err != nil {
		return nil, err
	}

	typ := keyTypeToRPC(key.Type())

	return &Key{
		ID:    key.ID().String(),
		User:  userResultToRPC(result),
		Type:  typ,
		Saved: saved,
	}, nil
}

// KeyRemove (RPC) removes a key.
func (s *service) KeyRemove(ctx context.Context, req *KeyRemoveRequest) (*KeyRemoveResponse, error) {
	if req.KID == "" {
		return nil, errors.Errorf("kid not specified")
	}
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}
	ok, err := s.ks.Delete(kid)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, keys.NewErrNotFound(kid.String())
	}

	_, err = s.scs.DeleteSigchain(kid)
	if err != nil {
		return nil, err
	}

	if _, err := s.users.Update(ctx, kid); err != nil {
		return nil, err
	}

	return &KeyRemoveResponse{}, nil
}

// KeyGenerate (RPC) creates a key.
func (s *service) KeyGenerate(ctx context.Context, req *KeyGenerateRequest) (*KeyGenerateResponse, error) {
	if req.Type == UnknownKeyType {
		return nil, errors.Errorf("no key type specified")
	}
	var kid keys.ID
	switch req.Type {
	case Ed25519:
		key := keys.GenerateEd25519Key()
		if err := s.ks.SaveSignKey(key); err != nil {
			return nil, err
		}
		kid = key.ID()
	case Curve25519:
		key := keys.GenerateCurve25519Key()
		if err := s.ks.SaveBoxKey(key); err != nil {
			return nil, err
		}
		kid = key.ID()
	default:
		return nil, errors.Errorf("unknown key type %s", req.Type)
	}

	return &KeyGenerateResponse{
		KID: kid.String(),
	}, nil
}

func (s *service) parseKID(kid string) (keys.ID, error) {
	if kid == "" {
		return "", errors.Errorf("no kid specified")
	}
	id, err := keys.ParseID(kid)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *service) parseKey(kid string, required bool) (keys.Key, error) {
	if kid == "" {
		if required {
			return nil, errors.Errorf("no kid specified")
		}
		return nil, nil
	}
	id, err := keys.ParseID(kid)
	if err != nil {
		return nil, err
	}
	key, err := s.ks.Key(id)
	if err != nil {
		return nil, err
	}
	if key == nil && required {
		return nil, keys.NewErrNotFound(kid)
	}
	return key, nil
}

func (s *service) parseSignKey(kid string, required bool) (*keys.SignKey, error) {
	if kid == "" {
		if required {
			return nil, errors.Errorf("no kid specified")
		}
		return nil, nil
	}
	id, err := keys.ParseID(kid)
	if err != nil {
		return nil, err
	}
	switch id.KeyType() {
	case keys.Ed25519Public:
		key, err := s.ks.SignKey(id)
		if err != nil {
			return nil, err
		}
		if key == nil && required {
			return nil, keys.NewErrNotFound(kid)
		}
		return key, nil
	default:
		return nil, errors.Errorf("unsupported key type for signing %s", id)
	}
}

func (s *service) parseBoxKey(kid string, required bool) (*keys.BoxKey, error) {
	if kid == "" {
		if required {
			return nil, errors.Errorf("no kid specified")
		}
		return nil, nil
	}
	id, err := keys.ParseID(kid)
	if err != nil {
		return nil, err
	}

	switch id.KeyType() {
	case keys.Ed25519Public:
		key, err := s.ks.SignKey(id)
		if err != nil {
			return nil, err
		}
		if key == nil && required {
			return nil, keys.NewErrNotFound(kid)
		}
		return key.Curve25519Key(), nil
	case keys.Curve25519Public:
		key, err := s.ks.BoxKey(id)
		if err != nil {
			return nil, err
		}
		if key == nil && required {
			return nil, keys.NewErrNotFound(kid)
		}
		return key, nil
	default:
		return nil, errors.Errorf("unsupported key type for %s", id)
	}
}

// checkSenderID checks if the ID is a box public key, finds the sign public key
// equivalent if found and returns that ID, otherwise returns itself.
func (s *service) checkSenderID(id keys.ID) (keys.ID, error) {
	if id.IsCurve25519() {
		bpk, err := keys.BoxPublicKeyForID(id)
		if err != nil {
			return "", err
		}
		spk, err := s.ks.FindEd25519PublicKey(bpk)
		if err != nil {
			return "", err
		}
		if spk == nil {
			// Not found, return original id.
			return id, nil
		}
		return spk.ID(), nil
	}
	return id, nil
}
