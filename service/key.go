package service

import (
	"context"
	strings "strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

// Key (RPC) ...
func (s *service) Key(ctx context.Context, req *KeyRequest) (*KeyResponse, error) {
	var kid keys.ID
	if req.User != "" {
		usr, err := s.searchUserLocalExact(ctx, req.User)
		if err != nil {
			return nil, err
		}
		if usr == nil {
			return &KeyResponse{}, nil
		}
		kid = usr.User.KID
	} else {
		k, err := s.parseKIDOrCurrent(req.KID)
		if err != nil {
			return nil, err
		}
		kid = k
	}

	key, err := s.key(ctx, kid)
	if err != nil {
		return nil, err
	}

	return &KeyResponse{
		Key: key,
	}, nil
}

// Emoji for KeyType.
func (t KeyType) Emoji() string {
	switch t {
	case PrivateKeyType:
		return "🔑" // 🔐
	case PublicKeyType:
		return "🖋️" // 🔏
	default:
		return "❓"
	}
}

func (s *service) key(ctx context.Context, kid keys.ID) (*Key, error) {
	logger.Debugf("Loading key %s", kid)
	if s.db == nil {
		return nil, errors.Errorf("db is locked")
	}

	typ := PublicKeyType
	var users []*User
	saved := false

	key, err := s.loadKey(kid)
	if err != nil {
		return nil, err
	}
	if key != nil {
		saved = true
		typ = PrivateKeyType
	} else {
		typ = PublicKeyType
	}

	res, err := s.users.Get(ctx, kid)
	if err != nil {
		return nil, err
	}
	users = userResultsToRPC(res)

	return &Key{
		ID:    kid.String(),
		Users: users,
		Type:  typ,
		Saved: saved,
	}, nil
}

// KeyBackup (RPC) returns a seed phrase which can be used to recover a key.
func (s *service) KeyBackup(ctx context.Context, req *KeyBackupRequest) (*KeyBackupResponse, error) {
	if req.KID == "" {
		return nil, errors.Errorf("no KID specified")
	}
	key, err := s.parseKey(req.KID)
	if err != nil {
		return nil, err
	}
	seedPhrase, err := keys.BytesToPhrase(key.Seed())
	if err != nil {
		return nil, err
	}
	return &KeyBackupResponse{
		SeedPhrase: seedPhrase,
	}, nil
}

// KeyRecover (RPC) recovers a key from a recovery (seed) phrase.
func (s *service) KeyRecover(ctx context.Context, req *KeyRecoverRequest) (*KeyRecoverResponse, error) {
	if req.SeedPhrase == "" {
		return nil, errors.Errorf("no seed phrase specified")
	}
	if !keys.IsValidPhrase(req.SeedPhrase, true) {
		return nil, errors.Errorf("invalid recovery phrase")
	}

	b, err := keys.PhraseToBytes(req.SeedPhrase, true)
	if err != nil {
		return nil, err
	}

	key, err := keys.NewSignKeyFromSeed(b)
	if err != nil {
		return nil, err
	}

	existing, err := s.ks.SignKey(key.ID())
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.Errorf("key already exists")
	}

	if err := s.saveKey(key); err != nil {
		return nil, err
	}

	return &KeyRecoverResponse{
		KID: key.ID().String(),
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
	key, err := s.loadKey(kid)
	if err != nil {
		return nil, err
	}
	if key != nil {
		if err := s.ensureNotAuthKey(key.ID()); err != nil {
			return nil, err
		}
		kid := key.ID()
		seedPhrase := strings.TrimSpace(req.SeedPhrase)

		if seedPhrase == "" {
			return nil, errors.Errorf("seed-phrase is required to remove a key, use `keys backup` to get the seed phrase")
		}

		keySeedPhrase, err := keys.BytesToPhrase(key.Seed())
		if err != nil {
			return nil, err
		}
		if seedPhrase != keySeedPhrase {
			return nil, errors.Errorf("seed phrase doesn't match")
		}

		ok, err := s.ks.Delete(kid.String())
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, keys.NewErrNotFound(kid.String())
		}
	}

	ok, err := s.scs.DeleteSigchain(kid)
	if err != nil {
		return nil, err
	}

	if key == nil && !ok {
		return nil, keys.NewErrNotFound(kid.String())
	}

	return &KeyRemoveResponse{}, nil
}

// KeyGenerate (RPC) creates a key.
func (s *service) KeyGenerate(ctx context.Context, req *KeyGenerateRequest) (*KeyGenerateResponse, error) {
	key := keys.GenerateSignKey()

	if err := s.saveKey(key); err != nil {
		return nil, err
	}

	return &KeyGenerateResponse{
		KID: key.ID().String(),
	}, nil
}

func (s *service) loadKIDs(all bool) ([]keys.ID, error) {
	ks, err := s.loadKeys()
	if err != nil {
		return nil, err
	}
	kids := keys.NewIDSet()
	for _, k := range ks {
		kids.Add(k.ID())
	}
	if all {
		pkids, err := s.scs.KIDs()
		if err != nil {
			return nil, err
		}
		kids.AddAll(pkids)
	}
	return kids.IDs(), nil
}

func (s *service) loadKeys() ([]*keys.SignKey, error) {
	return s.ks.SignKeys()
}

func (s *service) isAuthKey(id keys.ID) (bool, error) {
	item, err := s.ks.Get(id.String())
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, keys.NewErrNotFound(id.String())
	}
	auth := item.SecretFor("auth").String() == "1"
	return auth, nil
}

func (s *service) ensureAuthKey(id keys.ID) error {
	auth, err := s.isAuthKey(id)
	if err != nil {
		return err
	}
	if !auth {
		return errors.Errorf("expected an auth key")
	}
	return nil
}

func (s *service) ensureNotAuthKey(id keys.ID) error {
	auth, err := s.isAuthKey(id)
	if err != nil {
		return err
	}
	if auth {
		return errors.Errorf("expected a private key (not an auth key)")
	}
	return nil
}

func (s *service) loadKey(id keys.ID) (*keys.SignKey, error) {
	return s.ks.SignKey(id)
}

func (s *service) saveKey(key *keys.SignKey) error {
	item := keys.NewSignKeyItem(key)
	return s.ks.Set(item)
}

func (s *service) setCurrentKey(key *keys.SignKey) error {
	link := keyring.NewItem(".current", keyring.NewStringSecret(key.ID().String()), "")
	if err := s.ks.Set(link); err != nil {
		return err
	}
	return nil
}

func (s *service) loadCurrentKey() (*keys.SignKey, error) {
	item, err := s.ks.Get(".current")
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	id, err := keys.ParseID(string(item.SecretData()))
	if err != nil {
		return nil, err
	}
	return s.loadKey(id)
}

func (s *service) parseKIDOrCurrent(id string) (keys.ID, error) {
	if id == "" {
		key, err := s.loadCurrentKey()
		if err != nil {
			return "", err
		}
		if key == nil {
			return "", errors.Errorf("no kid specified")
		}
		return key.ID(), nil
	}
	kid, err := keys.ParseID(id)
	if err != nil {
		return "", err
	}
	return kid, nil
}

func (s *service) parseKeyOrCurrent(id string) (*keys.SignKey, error) {
	if id == "" {
		key, err := s.loadCurrentKey()
		if err != nil {
			return nil, err
		}
		if key == nil {
			return nil, errors.Errorf("no kid specified")
		}
		return key, nil
	}
	return s.parseKey(id)
}

func (s *service) parseKey(id string) (*keys.SignKey, error) {
	if id == "" {
		return nil, errors.Errorf("no kid specified")
	}
	kid, err := keys.ParseID(id)
	if err != nil {
		return nil, err
	}
	key, err := s.loadKey(kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, keys.NewErrNotFound(kid.String())
	}
	return key, nil
}
