package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// Key (RPC) ...
func (s *service) Key(ctx context.Context, req *KeyRequest) (*KeyResponse, error) {
	kid, err := s.parseIdentity(context.TODO(), req.Identity, false)
	if err != nil {
		return nil, err
	}

	if req.Update {
		if _, _, err := s.update(ctx, kid); err != nil {
			return nil, err
		}
	} else {
		if err := s.checkForKeyUpdate(ctx, kid, false); err != nil {
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

func (s *service) verifyKey(ctx context.Context, kid keys.ID) (*Key, error) {
	if err := s.ensureVerified(ctx, kid); err != nil {
		return nil, err
	}
	return s.loadKey(ctx, kid)
}

func (s *service) loadKey(ctx context.Context, kid keys.ID) (*Key, error) {
	key, err := s.vault.Key(kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return s.keyIDToRPC(ctx, kid)
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
	typ := keyTypeToRPC(key.Type())
	out := &Key{
		ID:    key.ID().String(),
		Type:  typ,
		Saved: true,
	}

	if err := s.fillKey(ctx, key.ID(), out); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *service) keyIDToRPC(ctx context.Context, kid keys.ID) (*Key, error) {
	key, err := s.vault.Key(kid)
	if err != nil {
		return nil, err
	}
	if key != nil {
		return s.keyToRPC(ctx, key)
	}

	typ := keyTypeToRPC(kid.PublicKeyType())

	out := &Key{
		ID:    kid.String(),
		Type:  typ,
		Saved: false,
	}
	if err := s.fillKey(ctx, kid, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *service) fillKey(ctx context.Context, kid keys.ID, key *Key) error {
	res, err := s.users.Get(ctx, kid)
	if err != nil {
		return err
	}

	key.User = userResultToRPC(res)

	// Sigchain info
	sc, err := s.scs.Sigchain(kid)
	if err != nil {
		return err
	}
	if sc != nil {
		key.SigchainLength = int32(sc.Length())
		last := sc.Last()
		if last != nil {
			key.SigchainUpdatedAt = tsutil.Millis(last.Timestamp)
		}
	}
	return nil
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
	ok, err := s.vault.Delete(kid.String())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, keys.NewErrNotFound(kid.String())
	}

	if kid.IsEdX25519() {
		_, err = s.scs.DeleteSigchain(kid)
		if err != nil {
			return nil, err
		}
		if _, err := s.users.Update(ctx, kid); err != nil {
			return nil, err
		}
	}

	return &KeyRemoveResponse{}, nil
}

// KeyGenerate (RPC) creates a key.
func (s *service) KeyGenerate(ctx context.Context, req *KeyGenerateRequest) (*KeyGenerateResponse, error) {
	if req.Type == UnknownKeyType {
		return nil, errors.Errorf("no key type specified")
	}
	var key keys.Key
	switch req.Type {
	case EdX25519:
		key = keys.GenerateEdX25519Key()
	case X25519:
		key = keys.GenerateX25519Key()
	default:
		return nil, errors.Errorf("unknown key type %s", req.Type)
	}
	if err := s.vault.SaveKey(key); err != nil {
		return nil, err
	}
	kid := key.ID()

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
	key, err := s.vault.Key(id)
	if err != nil {
		return nil, err
	}
	if key == nil && required {
		return nil, keys.NewErrNotFound(kid)
	}
	return key, nil
}

// convertKey checks if the ID is a X25519 public key, finds the EdX25519 public key
// equivalent if found, otherwise returns itself.
func (s *service) convertX25519ID(kid keys.ID) (keys.ID, error) {
	if kid == "" {
		return "", nil
	}
	if kid.IsX25519() {
		logger.Debugf("Convert sender %s", kid)
		spk, err := s.vault.EdX25519PublicKey(kid)
		if err != nil {
			return "", err
		}
		if spk == nil {
			logger.Debugf("No edx25519 key found for %s", kid)
			// Not found, return original id.
			return kid, nil
		}
		logger.Debugf("Found edx25519 key %s (for %s)", spk, kid)
		return spk.ID(), nil
	}
	return kid, nil
}

func (s *service) parseSigner(signer string, required bool) (*keys.EdX25519Key, error) {
	if signer == "" {
		if required {
			return nil, errors.Errorf("no signer specified")
		}
		return nil, nil
	}
	kid, err := s.parseIdentity(context.TODO(), signer, false)
	if err != nil {
		return nil, err
	}
	key, err := s.vault.Key(kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, keys.NewErrNotFound(kid.String())
	}
	sk, ok := key.(*keys.EdX25519Key)
	if !ok {
		return nil, errors.Errorf("unsupported key type for %s", kid)
	}
	return sk, nil
}

func (s *service) parseBoxKey(kid keys.ID) (*keys.X25519Key, error) {
	if kid == "" {
		return nil, nil
	}
	key, err := s.vault.Key(kid)
	if err != nil {
		return nil, err
	}
	switch k := key.(type) {
	case *keys.EdX25519Key:
		return k.X25519Key(), nil
	case *keys.X25519Key:
		return k, nil
	default:
		return nil, errors.Errorf("unsupported key type for %s", kid)
	}
}

func (s *service) parseSignKey(id string, required bool) (*keys.EdX25519Key, error) {
	if id == "" {
		if required {
			return nil, errors.Errorf("no kid specified")
		}
		return nil, nil
	}
	kid, err := keys.ParseID(id)
	if err != nil {
		return nil, err
	}
	key, err := s.vault.Key(kid)
	if err != nil {
		return nil, err
	}
	if key == nil && required {
		return nil, keys.NewErrNotFound(id)
	}
	switch k := key.(type) {
	case *keys.EdX25519Key:
		return k, nil
	default:
		return nil, errors.Errorf("unsupported key type for %s", kid)
	}
}

func (s *service) parseIdentityForEdX25519Key(ctx context.Context, identity string) (*keys.EdX25519Key, error) {
	kid, err := s.parseIdentity(ctx, identity, false)
	if err != nil {
		return nil, err
	}
	key, err := s.vault.Key(kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, keys.NewErrNotFound(kid.String())
	}
	switch k := key.(type) {
	case *keys.EdX25519Key:
		return k, nil
	default:
		return nil, errors.Errorf("unsupported key type for %s", kid)
	}
}

func (s *service) parseIdentityForEdX25519PublicKey(ctx context.Context, identity string) (*keys.EdX25519PublicKey, error) {
	kid, err := s.parseIdentity(ctx, identity, false)
	if err != nil {
		return nil, err
	}
	return keys.NewEdX25519PublicKeyFromID(kid)
}
