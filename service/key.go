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
		usr, err := s.searchUserLocalExact(ctx, req.User)
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

func (s *service) loadKey(ctx context.Context, id keys.ID) (*Key, error) {
	hrp, _, err := id.Decode()
	if err != nil {
		return nil, err
	}
	switch hrp {
	case keys.SignKeyType:
		sk, err := s.ks.SignKey(id)
		if err != nil {
			return nil, err
		}
		if sk != nil {
			return s.signKeyToRPC(ctx, sk)
		}
		spk, err := s.ks.SignPublicKey(id)
		if err != nil {
			return nil, err
		}
		if spk != nil {
			return s.signPublicKeyToRPC(ctx, spk)
		}
		return s.signKeyIDToRPC(ctx, id)
	default:
		return nil, errors.Errorf("unrecognized key type %s", hrp)
	}

}

func (s *service) signKeyToRPC(ctx context.Context, sk *keys.SignKey) (*Key, error) {
	users, err := s.users.Get(ctx, sk.ID())
	if err != nil {
		return nil, err
	}

	return &Key{
		ID:    sk.ID().String(),
		Users: userResultsToRPC(users),
		Type:  PrivateKeyType,
		Saved: true,
	}, nil
}

func (s *service) signPublicKeyToRPC(ctx context.Context, spk *keys.SignPublicKey) (*Key, error) {
	users, err := s.users.Get(ctx, spk.ID())
	if err != nil {
		return nil, err
	}

	return &Key{
		ID:    spk.ID().String(),
		Users: userResultsToRPC(users),
		Type:  PublicKeyType,
		Saved: true,
	}, nil
}

func (s *service) signKeyIDToRPC(ctx context.Context, id keys.ID) (*Key, error) {
	users, err := s.users.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Key{
		ID:    id.String(),
		Users: userResultsToRPC(users),
		Type:  PublicKeyType,
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
	key := keys.GenerateSignKey()

	if err := s.ks.SaveSignKey(key); err != nil {
		return nil, err
	}

	return &KeyGenerateResponse{
		KID: key.ID().String(),
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

func (s *service) parseKey(kid string) (*keys.SignKey, error) {
	if kid == "" {
		return nil, errors.Errorf("no kid specified")
	}
	id, err := keys.ParseID(kid)
	if err != nil {
		return nil, err
	}
	key, err := s.ks.SignKey(id)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, keys.NewErrNotFound(kid)
	}
	return key, nil
}
