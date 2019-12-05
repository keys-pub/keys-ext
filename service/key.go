package service

import (
	"context"
	strings "strings"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

// Key (RPC) ...
func (s *service) Key(ctx context.Context, req *KeyRequest) (*KeyResponse, error) {
	var kid keys.ID
	if req.KID != "" {
		k, err := keys.ParseID(req.KID)
		if err != nil {
			return nil, err
		}
		kid = k
	} else if req.User != "" {
		usr, err := s.findUserByName(ctx, req.User)
		if err != nil {
			return nil, err
		}
		if usr == nil {
			return &KeyResponse{}, nil
		}
		kid = usr.KID
	} else {
		return nil, errors.Errorf("no kid or user specified")
	}

	keyOut, err := s.key(ctx, kid, req.Check, req.Update)
	if err != nil {
		return nil, err
	}
	return &KeyResponse{
		Key: keyOut,
	}, nil
}

// Emoji for KeyType.
func (t KeyType) Emoji() string {
	switch t {
	case PrivateKeyType:
		return "🔑"
	case PublicKeyType:
		return "✉️"
	default:
		return "❓"
	}
}

func (s *service) key(ctx context.Context, kid keys.ID, check bool, update bool) (*Key, error) {
	logger.Debugf("Loading key %s", kid)

	typ := PublicKeyType
	var usrs []*keys.User
	saved := false
	var createdAt time.Time
	var publishedAt time.Time
	var savedAt time.Time
	var updatedAt time.Time

	sc, err := s.scs.Sigchain(kid)
	if err != nil {
		return nil, err
	}
	if sc != nil {
		logger.Debugf("Found local sigchain")
		saved = true

		if update {
			ok, err := s.pull(ctx, kid)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, keys.NewErrNotFound(kid, keys.SigchainType)
			}
		}

		sts := sc.Statements()
		if len(sts) > 0 {
			st := sts[0]
			res, err := s.loadResource(ctx, st.KeyPath())
			if err != nil {
				return nil, err
			}
			if res != nil {
				logger.Debugf("Found local resource %+v", res)
				publishedAt = res.Metadata.CreatedAt
			}

			// Lookup when it was saved
			doc, err := s.db.Get(ctx, st.KeyPath())
			if err != nil {
				return nil, err
			}
			if doc != nil {
				savedAt = doc.CreatedAt
				updatedAt = doc.UpdatedAt
			}
		}
	} else {
		logger.Debugf("Loading remote sigchain %s", kid)
		resp, err := s.remote.Sigchain(kid)
		if err != nil {
			return nil, err
		}
		sc, err = resp.Sigchain()
		if err != nil {
			return nil, err
		}
		sts := sc.Statements()
		if len(sts) > 0 {
			st := sts[0]
			publishedAt = resp.MetadataFor(st).CreatedAt
		}
	}

	if sc != nil {
		sts := sc.Statements()
		if len(sts) > 0 {
			st := sts[0]
			createdAt = st.Timestamp
		}
	}

	if sc != nil {
		if check {
			_, usrsErr := keys.UserCheck(ctx, sc, nil, time.Now)
			if usrsErr != nil {
				return nil, usrsErr
			}
		}
		usrs = sc.Users()
	}

	key, err := s.loadPrivateKey(kid)
	if err != nil {
		return nil, err
	}
	if key != nil {
		saved = true
		typ = PrivateKeyType
	}

	if key == nil && sc == nil {
		return nil, keys.NewErrNotFound(kid, keys.KeyType)
	}

	return &Key{
		KID:         kid.String(),
		Users:       usersToRPC(usrs),
		Type:        typ,
		Saved:       saved,
		CreatedAt:   int64(keys.TimeToMillis(createdAt)),
		PublishedAt: int64(keys.TimeToMillis(publishedAt)),
		SavedAt:     int64(keys.TimeToMillis(savedAt)),
		UpdatedAt:   int64(keys.TimeToMillis(updatedAt)),
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
	seedPhrase := keys.SeedPhrase(key)
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

	key, err := keys.NewKeyFromSeedPhrase(req.SeedPhrase, true)
	if err != nil {
		return nil, err
	}

	existing, err := s.ks.Key(key.ID())
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.Errorf("key already exists")
	}

	sc, err := s.scs.Sigchain(key.ID())
	if err != nil {
		return nil, err
	}
	generateSigchain := true
	if sc != nil {
		generateSigchain = false
	}

	if err := s.saveKey(key, false, generateSigchain); err != nil {
		return nil, err
	}

	if req.PublishPublicKey {
		if _, err := s.push(key.ID()); err != nil {
			return nil, err
		}
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
	key, err := s.loadPrivateKey(kid)
	if err != nil {
		return nil, err
	}
	if key != nil {
		if err := s.ensureNotAuthKey(key.ID()); err != nil {
			return nil, err
		}
		kid := key.ID()

		// If current key, clear it
		ck, ckErr := s.loadCurrentKey()
		if ckErr != nil {
			return nil, ckErr
		}
		if ck != nil && ck.ID() == kid {
			if err := s.clearCurrentKey(); err != nil {
				return nil, err
			}
		}

		seedPhrase := strings.TrimSpace(req.SeedPhrase)

		if seedPhrase == "" {
			return nil, errors.Errorf("seed-phrase is required to remove a key, use `keys backup` to get the seed phrase")
		}

		keySeedPhrase := keys.SeedPhrase(key)
		if seedPhrase != keySeedPhrase {
			return nil, errors.Errorf("seed phrase doesn't match")
		}

		ok, deleteErr := s.ks.Delete(kid.String())
		if deleteErr != nil {
			return nil, deleteErr
		}
		if !ok {
			return nil, keys.NewErrNotFound(kid, keys.KeyType)
		}
	}

	ok, err := s.scs.DeleteSigchain(kid)
	if err != nil {
		return nil, err
	}

	if key == nil && !ok {
		return nil, keys.NewErrNotFound(kid, keys.KeyType)
	}

	return &KeyRemoveResponse{}, nil
}

// KeyGenerate (RPC) creates a key.
func (s *service) KeyGenerate(ctx context.Context, req *KeyGenerateRequest) (*KeyGenerateResponse, error) {
	key := keys.GenerateKey()

	if err := s.saveKey(key, false, true); err != nil {
		return nil, err
	}

	if req.PublishPublicKey {
		_, pushErr := s.push(key.ID())
		if pushErr != nil {
			return nil, pushErr
		}
	}

	return &KeyGenerateResponse{
		KID: key.ID().String(),
	}, nil
}

func (s *service) kidsSet(all bool) (*keys.IDSet, error) {
	ks, ksErr := s.loadPrivateKeys()
	if ksErr != nil {
		return nil, ksErr
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
	return kids, nil
}

func (s *service) loadPrivateKeys() ([]keys.Key, error) {
	items, lierr := s.ks.Keyring().List(&keyring.ListOpts{
		Type: keys.KeyType,
	})
	if lierr != nil {
		return nil, lierr
	}
	out := make([]keys.Key, 0, len(items))
	for _, item := range items {
		key, err := keys.AsKey(item)
		if err != nil {
			return nil, err
		}
		out = append(out, key)
	}
	return out, nil
}

func (s *service) isAuthKey(id keys.ID) (bool, error) {
	item, err := s.ks.Get(id)
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, keys.NewErrNotFound(id, keys.KeyType)
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

func (s *service) loadPrivateKey(id keys.ID) (keys.Key, error) {
	item, err := s.ks.Get(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	key, err := keys.AsKey(item)
	if err != nil {
		return nil, err
	}
	// auth := item.SecretFor("auth").String() == "1"
	return key, nil
}

func (s *service) saveKey(key keys.Key, auth bool, generateSigchain bool) error {
	if generateSigchain {
		if s.scs == nil {
			return errors.Errorf("no sigchain store set")
		}
		if err := s.scs.SaveSigchain(keys.GenerateSigchain(key, s.Now())); err != nil {
			return err
		}
	}
	// Save to keyring
	item := keys.NewKeyItem(key)
	if auth {
		item.SetSecretFor("auth", keyring.NewStringSecret("1"))
	}
	if err := s.ks.Set(item); err != nil {
		return err
	}

	// Set default key if none set and we generated our first key
	ck, ckErr := s.loadCurrentKey()
	if ckErr != nil {
		return ckErr
	}
	ks, ksErr := s.ks.Keys()
	if ksErr != nil {
		return ksErr
	}
	if ck == nil && len(ks) == 1 && auth {
		if err := s.setCurrentKey(ks[0].ID()); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) loadCurrentKey() (keys.Key, error) {
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
	return s.loadPrivateKey(id)
}

func (s *service) setCurrentKey(kid keys.ID) error {
	if err := s.ensureAuthKey(kid); err != nil {
		return err
	}
	return s.ks.Set(keyring.NewItem(".current", keyring.NewStringSecret(kid.String()), ""))
}

func (s *service) clearCurrentKey() error {
	_, err := s.ks.Delete(".current")
	return err
}

func recipientIDs(pks []keys.PublicKey) []keys.ID {
	ids := make([]keys.ID, 0, len(pks))
	for _, pk := range pks {
		ids = append(ids, pk.ID())
	}
	return ids
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

func (s *service) parseKeyOrCurrent(id string) (keys.Key, error) {
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

func (s *service) parseKey(id string) (keys.Key, error) {
	if id == "" {
		return nil, errors.Errorf("no kid specified")
	}
	kid, err := keys.ParseID(id)
	if err != nil {
		return nil, err
	}
	key, err := s.loadPrivateKey(kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, keys.NewErrNotFound(kid, keys.KeyType)
	}
	return key, nil
}

func (s *service) keyRetrieve(recipient keys.Key, id keys.ID) (keys.Key, error) {
	key, err := s.ks.Key(id)
	if err != nil {
		return nil, err
	}
	if key != nil {
		return nil, errors.Errorf("already have key")
	}
	if s.remote == nil {
		return nil, errors.Errorf("no remote set")
	}
	b, sharedErr := s.remote.Shared(recipient, id)
	if sharedErr != nil {
		return nil, sharedErr
	}
	if b == nil {
		return nil, keys.NewErrNotFound(id, keys.KeyType)
	}
	if len(b) != 32 {
		return nil, errors.Errorf("invalid shared key bytes")
	}
	key, err = keys.NewKey(keys.Bytes32(b))
	if err != nil {
		return nil, err
	}
	if key.ID() != id {
		return nil, errors.Errorf("invalid shared key id")
	}
	if err := s.ks.SaveKey(key, false, s.Now()); err != nil {
		return nil, err
	}
	return key, nil
}

func (s *service) keyShare(recipient keys.ID, key keys.Key) error {
	if s.remote == nil {
		return errors.Errorf("no remote set")
	}
	pk, err := s.ks.PublicKey(recipient)
	if err != nil {
		return err
	}
	if pk == nil {
		return keys.NewErrNotFound(recipient, keys.PublicKeyType)
	}
	_, shareErr := s.remote.Share(pk, key, key.Seed()[:])
	if shareErr != nil {
		return shareErr
	}
	return nil
}

// func (s *service) parseShareKey(id string, recipient keys.Key) (keys.Key, error) {
// 	kid, err := keys.ParseID(id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	key, err := s.ks.Key(kid)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if key != nil {
// 		return key, nil
// 	}
// 	return s.keyRetrieve(recipient, kid)
// }

func (s *service) parsePublicKeys(strs string) ([]keys.PublicKey, error) {
	ids, err := keys.ParseIDs(strings.Split(strs, ","))
	if err != nil {
		return nil, err
	}
	pks := make([]keys.PublicKey, 0, len(ids))
	for _, id := range ids {
		pk, err := s.ks.PublicKey(id)
		if err != nil {
			return nil, err
		}
		if pk == nil {
			return nil, keys.NewErrNotFound(id, keys.PublicKeyType)
		}
		pks = append(pks, pk)
	}
	return pks, nil
}

// KeyShare (RPC) ...
func (s *service) KeyShare(ctx context.Context, req *KeyShareRequest) (*KeyShareResponse, error) {
	if req.KID == "" {
		return nil, errors.Errorf("no kid specified")
	}
	key, err := s.parseKey(req.KID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNotAuthKey(key.ID()); err != nil {
		return nil, err
	}
	if req.Recipient == "" {
		return nil, errors.Errorf("no recipient specified")
	}
	recipient, err := keys.ParseID(req.Recipient)
	if err != nil {
		return nil, err
	}

	if _, err := s.pull(ctx, recipient); err != nil {
		return nil, err
	}
	if _, err := s.push(key.ID()); err != nil {
		return nil, err
	}

	if err := s.keyShare(recipient, key); err != nil {
		return nil, err
	}
	return &KeyShareResponse{}, nil
}

// KeyRetrieve (RPC) ...
func (s *service) KeyRetrieve(ctx context.Context, req *KeyRetrieveRequest) (*KeyRetrieveResponse, error) {
	if req.KID == "" {
		return nil, errors.Errorf("no kid specified")
	}
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}
	if req.Recipient == "" {
		return nil, errors.Errorf("no recipient specified")
	}
	recipient, err := s.parseKey(req.Recipient)
	if err != nil {
		return nil, err
	}

	if _, err := s.pull(ctx, kid); err != nil {
		return nil, err
	}
	if _, err := s.pull(ctx, recipient.ID()); err != nil {
		return nil, err
	}

	if _, err := s.keyRetrieve(recipient, kid); err != nil {
		return nil, err
	}

	return &KeyRetrieveResponse{}, nil
}
