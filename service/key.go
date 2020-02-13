package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Key (RPC) ...
func (s *service) Key(ctx context.Context, req *KeyRequest) (*KeyResponse, error) {
	kid, err := s.parseIdentity(context.TODO(), req.Identity)
	if err != nil {
		return nil, err
	}

	if req.Update {
		if _, err := s.update(ctx, kid); err != nil {
			return nil, err
		}
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
	case keys.EdX25519:
		return "🖋️"
	case keys.EdX25519Public:
		return "🖋️"
	case keys.X25519:
		return "🔑"
	case keys.X25519Public:
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
		return s.keyIDToRPC(ctx, id)
	}
	return s.keyToRPC(ctx, key)
}

var keyTypeStrings = []string{
	string(keys.EdX25519),
	string(keys.EdX25519Public),
	string(keys.X25519),
	string(keys.X25519Public),
}

func parseKeyType(s string) (KeyType, error) {
	switch s {
	case string(keys.EdX25519):
		return EdX25519, nil
	case string(keys.EdX25519Public):
		return EdX25519Public, nil
	case string(keys.X25519):
		return X25519, nil
	case string(keys.X25519Public):
		return X25519Public, nil
	default:
		return UnknownKeyType, errors.Errorf("unsupported key type %s", s)
	}
}

func keyTypeFromRPC(t KeyType) (keys.KeyType, error) {
	switch t {
	case EdX25519:
		return keys.EdX25519, nil
	case EdX25519Public:
		return keys.EdX25519Public, nil
	case X25519:
		return keys.X25519, nil
	case X25519Public:
		return keys.X25519Public, nil
	default:
		return "", errors.Errorf("unsupported key type")
	}
}

func keyTypeToRPC(t keys.KeyType) KeyType {
	switch t {
	case keys.EdX25519:
		return EdX25519
	case keys.EdX25519Public:
		return EdX25519Public
	case keys.X25519:
		return X25519
	case keys.X25519Public:
		return X25519Public
	default:
		return UnknownKeyType
	}
}

func (s *service) keyToRPC(ctx context.Context, key keys.Key) (*Key, error) {
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
		Saved: true,
	}, nil
}

func (s *service) keyIDToRPC(ctx context.Context, kid keys.ID) (*Key, error) {
	result, err := s.users.Get(ctx, kid)
	if err != nil {
		return nil, err
	}

	typ := keyTypeToRPC(kid.KeyType())

	return &Key{
		ID:    kid.String(),
		User:  userResultToRPC(result),
		Type:  typ,
		Saved: false,
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
	case EdX25519:
		key := keys.GenerateEdX25519Key()
		if err := s.ks.SaveSignKey(key); err != nil {
			return nil, err
		}
		kid = key.ID()
	case X25519:
		key := keys.GenerateX25519Key()
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

// checkSignerID checks if the ID is a box public key, finds the sign public key
// equivalent if found and returns that ID, otherwise returns itself.
func (s *service) checkSignerID(id keys.ID) (keys.ID, error) {
	if id.IsX25519() {
		bpk, err := keys.BoxPublicKeyForID(id)
		if err != nil {
			return "", err
		}
		spk, err := s.ks.FindEdX25519PublicKey(bpk)
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

func (s *service) parseSigner(signer string, required bool) (*keys.SignKey, error) {
	if signer == "" {
		if required {
			return nil, errors.Errorf("no signer specified")
		}
		return nil, nil
	}
	kid, err := s.parseIdentity(context.TODO(), signer)
	if err != nil {
		return nil, err
	}
	sk, err := s.ks.SignKey(kid)
	if err != nil {
		return nil, err
	}
	if sk == nil {
		return nil, keys.NewErrNotFound(kid.String())
	}
	return sk, nil
}

func (s *service) parseBoxKey(kid keys.ID) (*keys.BoxKey, error) {
	if kid == "" {
		return nil, nil
	}
	switch kid.KeyType() {
	case keys.EdX25519Public:
		key, err := s.ks.SignKey(kid)
		if err != nil {
			return nil, err
		}
		if key == nil {
			return nil, nil
		}
		return key.X25519Key(), nil
	case keys.X25519Public:
		return s.ks.BoxKey(kid)
	default:
		return nil, errors.Errorf("unsupported key type for %s", kid)
	}
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
	case keys.EdX25519Public:
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
