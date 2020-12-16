package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// Key (RPC) ...
func (s *service) Key(ctx context.Context, req *KeyRequest) (*KeyResponse, error) {
	kid, err := s.lookup(ctx, req.Key, &lookupOpts{SearchRemote: req.Search})
	if err != nil {
		return nil, err
	}

	if req.Update {
		if _, err := s.updateUser(ctx, kid); err != nil {
			return nil, err
		}
	} else {
		if err := s.checkForExpiredKey(ctx, kid); err != nil {
			return nil, err
		}
	}

	key, err := s.key(ctx, kid)
	if err != nil {
		return nil, err
	}

	return &KeyResponse{
		Key: key,
	}, nil
}

func (s *service) verifyKey(ctx context.Context, kid keys.ID) (*Key, error) {
	if err := s.ensureUserVerified(ctx, kid); err != nil {
		return nil, err
	}
	return s.key(ctx, kid)
}

func (s *service) keyToRPC(ctx context.Context, key *api.Key, saved bool) (*Key, error) {
	if key == nil {
		return nil, nil
	}
	out := &Key{
		ID:        key.ID.String(),
		Type:      key.Type,
		Saved:     saved,
		IsPrivate: len(key.Private) > 0,
	}

	if err := s.fillKey(ctx, key.ID, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *service) key(ctx context.Context, kid keys.ID) (*Key, error) {
	key, err := s.vault.Key(kid)
	if err != nil {
		return nil, err
	}
	if key != nil {
		return s.keyToRPC(ctx, key, true)
	}
	return s.keyToRPC(ctx, api.NewKey(kid), false)
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
		_, err = s.scs.Delete(kid)
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
	if req.Type == "" {
		return nil, errors.Errorf("no key type specified")
	}
	var key keys.Key
	switch req.Type {
	case string(keys.EdX25519):
		key = keys.GenerateEdX25519Key()
	case string(keys.X25519):
		key = keys.GenerateX25519Key()
	default:
		return nil, errors.Errorf("unknown key type %s", req.Type)
	}
	vk := api.NewKey(key)
	now := s.clock.NowMillis()
	vk.CreatedAt = now
	vk.UpdatedAt = now
	out, _, err := s.vault.SaveKey(vk)
	if err != nil {
		return nil, err
	}
	if err := s.scs.Index(out.ID); err != nil {
		return nil, err
	}

	return &KeyGenerateResponse{
		KID: out.ID.String(),
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

// convertIfX25519ID checks if the ID is a X25519 public key, finds the EdX25519 public key
// equivalent if found, otherwise returns itself.
func (s *service) convertIfX25519ID(kid keys.ID) (keys.ID, error) {
	if kid == "" {
		return "", nil
	}
	if kid.IsX25519() {
		logger.Debugf("Convert sender %s", kid)
		spk, err := s.vault.FindEdX25519PublicKey(kid)
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

func (s *service) edx25519Key(kid keys.ID) (*keys.EdX25519Key, error) {
	if kid == "" {
		return nil, nil
	}
	key, err := s.vault.Key(kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, keys.NewErrNotFound(kid.String())
	}
	sk := key.AsEdX25519()
	if sk == nil {
		return nil, keys.NewErrNotFound(kid.String())
	}
	return sk, nil
}

func (s *service) x25519Key(kid keys.ID) (*keys.X25519Key, error) {
	if kid == "" {
		return nil, nil
	}
	key, err := s.vault.Key(kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, keys.NewErrNotFound(kid.String())
	}
	bk := key.AsX25519()
	if bk == nil {
		return nil, keys.NewErrNotFound(kid.String())
	}
	return bk, nil
}

func (k *Key) userName() string {
	if k.User != nil && k.User.ID != "" {
		return k.User.ID
	}
	return k.ID
}

func keyUserNames(ks []*Key) []string {
	out := []string{}
	for _, k := range ks {
		out = append(out, k.userName())
	}
	return out
}
