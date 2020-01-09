package service

import (
	"context"

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

// KeyBackup (RPC) returns password protected key backup.
func (s *service) KeyBackup(ctx context.Context, req *KeyBackupRequest) (*KeyBackupResponse, error) {
	key, err := s.parseKeyOrCurrent(req.KID)
	if err != nil {
		return nil, err
	}
	keyBackup := seedToBackup(req.Password, key.Seed())
	return &KeyBackupResponse{
		KeyBackup: keyBackup,
	}, nil
}

// KeyRecover (RPC) recovers a key from a backup.
func (s *service) KeyRecover(ctx context.Context, req *KeyRecoverRequest) (*KeyRecoverResponse, error) {
	seed, err := backupToSeed(req.Password, req.KeyBackup)
	if err != nil {
		return nil, err
	}
	if len(seed) != 32 {
		return nil, errors.Errorf("invalid bytes")
	}

	key, err := keys.NewSignKeyFromSeed(keys.Bytes32(seed))
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

	if _, err := s.users.Update(ctx, kid); err != nil {
		return nil, err
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
