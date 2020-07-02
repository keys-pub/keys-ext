package vault

import (
	"bytes"
	"time"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Key for id.
func (v *Vault) Key(id keys.ID) (keys.Key, error) {
	item, err := v.Get(id.String())
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return KeyForItem(item)
}

// EdX25519Key for id.
func (v *Vault) EdX25519Key(id keys.ID) (*keys.EdX25519Key, error) {
	key, err := v.Key(id)
	if err != nil {
		return nil, err
	}
	sk, ok := key.(*keys.EdX25519Key)
	if !ok {
		return nil, nil
	}
	return sk, nil
}

// X25519Key for id.
func (v *Vault) X25519Key(id keys.ID) (*keys.X25519Key, error) {
	key, err := v.Key(id)
	if err != nil {
		return nil, err
	}
	bk, ok := key.(*keys.X25519Key)
	if !ok {
		return nil, nil
	}
	return bk, nil
}

// SaveKey saves a key.
func (v *Vault) SaveKey(key keys.Key) error {
	return v.Set(ItemForKey(key))
}

// Keys options.
var Keys = keysOpts{}

type keysOpts struct{}

func (s keysOpts) Types(t ...keys.KeyType) KeysOption {
	return func(o *KeysOptions) { o.Types = t }
}

// Keys in the vault.
func (v *Vault) Keys(opt ...KeysOption) ([]keys.Key, error) {
	opts := newKeysOptions(opt...)
	logger.Debugf("Keys...")

	items, err := v.Items()
	if err != nil {
		return nil, err
	}
	keys := make([]keys.Key, 0, len(items))
	for _, item := range items {
		// logger.Debugf("Key for item type: %s", item.Type)
		key, err := KeyForItem(item)
		if err != nil {
			return nil, err
		}
		key = filterKey(key, opts.Types)
		if key == nil {
			continue
		}
		keys = append(keys, key)
	}
	logger.Debugf("Found %d keys", len(keys))
	return keys, nil
}

func filterKey(key keys.Key, types []keys.KeyType) keys.Key {
	if key == nil || len(types) == 0 {
		return key
	}
	for _, t := range types {
		if t == key.Type() {
			return key
		}
	}
	return nil
}

// X25519Keys from the vault.
// Also includes edx25519 keys converted to x25519 keys.
func (v *Vault) X25519Keys() ([]*keys.X25519Key, error) {
	logger.Debugf("Listing x25519 keys...")
	items, err := v.Items()
	if err != nil {
		return nil, err
	}
	out := make([]*keys.X25519Key, 0, len(items))
	for _, item := range items {
		switch item.Type {
		case string(keys.X25519), string(keys.EdX25519):
			key, err := x25519KeyForItem(item)
			if err != nil {
				return nil, err
			}
			out = append(out, key)
		}
	}
	logger.Debugf("Found %d x25519 keys", len(out))
	return out, nil
}

// EdX25519Keys from the vault.
func (v *Vault) EdX25519Keys() ([]*keys.EdX25519Key, error) {
	items, err := v.Items()
	if err != nil {
		return nil, err
	}
	out := make([]*keys.EdX25519Key, 0, len(items))
	for _, item := range items {
		if item.Type != string(keys.EdX25519) {
			continue
		}
		key, err := edx25519KeyForItem(item)
		if err != nil {
			return nil, err
		}
		out = append(out, key)
	}
	return out, nil
}

// EdX25519PublicKeys from the vault.
// Includes public keys of EdX25519Key's.
func (v *Vault) EdX25519PublicKeys() ([]*keys.EdX25519PublicKey, error) {
	items, err := v.Items()
	if err != nil {
		return nil, err
	}
	out := make([]*keys.EdX25519PublicKey, 0, len(items))
	for _, item := range items {
		switch item.Type {
		case string(keys.EdX25519):
			key, err := edx25519KeyForItem(item)
			if err != nil {
				return nil, err
			}
			out = append(out, key.PublicKey())
		case string(keys.EdX25519Public):
			key, err := edx25519PublicKeyForItem(item)
			if err != nil {
				return nil, err
			}
			out = append(out, key)
		}
	}
	return out, nil
}

// EdX25519PublicKey searches all our EdX25519 public keys for a match to a converted
// X25519 public key.
func (v *Vault) EdX25519PublicKey(kid keys.ID) (*keys.EdX25519PublicKey, error) {
	logger.Debugf("Finding edx25519 key from an x25519 key %s", kid)
	spks, err := v.EdX25519PublicKeys()
	if err != nil {
		return nil, err
	}
	bpk, err := keys.NewX25519PublicKeyFromID(kid)
	if err != nil {
		return nil, err
	}
	for _, spk := range spks {
		if bytes.Equal(spk.X25519PublicKey().Bytes(), bpk.Bytes()) {
			logger.Debugf("Found ed25519 key %s", spk.ID())
			return spk, nil
		}
	}
	logger.Debugf("EdX25519 key not found")
	return nil, err
}

// ImportSaltpack imports key into the vault from a Saltpack message.
func (v *Vault) ImportSaltpack(msg string, password string, isHTML bool) (keys.Key, error) {
	key, err := keys.DecodeSaltpackKey(msg, password, isHTML)
	if err != nil {
		return nil, err
	}
	if err := v.Set(ItemForKey(key)); err != nil {
		return nil, err
	}
	return key, nil
}

// ExportSaltpack exports key from the vault to a Saltpack message.
func (v *Vault) ExportSaltpack(id keys.ID, password string) (string, error) {
	key, err := v.Key(id)
	if err != nil {
		return "", err
	}
	return keys.EncodeSaltpackKey(key, password)
}

// x25519KeyForItem returns a X25519Key for a vault Item.
// If item is a EdX25519Key it's converted to a X25519Key.
func x25519KeyForItem(item *Item) (*keys.X25519Key, error) {
	switch item.Type {
	case string(keys.X25519):
		bk := keys.NewX25519KeyFromPrivateKey(keys.Bytes32(item.Data))
		return bk, nil
	case string(keys.EdX25519):
		sk, err := edx25519KeyForItem(item)
		if err != nil {
			return nil, err
		}
		return sk.X25519Key(), nil
	default:
		return nil, errors.Errorf("item type %s != %s", item.Type, string(keys.X25519))
	}
}

// edx25519KeyForItem returns EdX25519Key for vault Item.
func edx25519KeyForItem(item *Item) (*keys.EdX25519Key, error) {
	if item.Type != string(keys.EdX25519) {
		return nil, errors.Errorf("item type %s != %s", item.Type, string(keys.EdX25519))
	}
	b := item.Data
	if len(b) != 64 {
		return nil, errors.Errorf("invalid number of bytes for ed25519 private key")
	}
	key := keys.NewEdX25519KeyFromPrivateKey(keys.Bytes64(b))
	return key, nil
}

// edx25519PublicKeyForItem returns EdX25519PublicKey for vault Item.
func edx25519PublicKeyForItem(item *Item) (*keys.EdX25519PublicKey, error) {
	switch item.Type {
	case string(keys.EdX25519Public):
		b := item.Data
		if len(b) != 32 {
			return nil, errors.Errorf("invalid number of bytes for ed25519 public key")
		}
		key := keys.NewEdX25519PublicKey(keys.Bytes32(b))
		return key, nil
	case string(keys.EdX25519):
		sk, err := edx25519KeyForItem(item)
		if err != nil {
			return nil, err
		}
		return sk.PublicKey(), nil
	default:
		return nil, errors.Errorf("invalid item type for edx25519 public key: %s", item.Type)
	}
}

// x25519PublicKeyForItem returns X25519PublicKey for vault Item.
func x25519PublicKeyForItem(item *Item) (*keys.X25519PublicKey, error) {
	switch item.Type {
	case string(keys.X25519Public):
		b := item.Data
		if len(b) != 32 {
			return nil, errors.Errorf("invalid number of bytes for x25519 public key")
		}
		key := keys.NewX25519PublicKey(keys.Bytes32(b))
		return key, nil
	case string(keys.X25519):
		bk, err := x25519KeyForItem(item)
		if err != nil {
			return nil, err
		}
		return bk.PublicKey(), nil
	default:
		return nil, errors.Errorf("invalid item type for x25519 public key: %s", item.Type)
	}
}

// ItemForKey returns vault.Item for a Key.
func ItemForKey(key keys.Key) *Item {
	return NewItem(key.ID().String(), key.Bytes(), string(key.Type()), time.Now())
}

// KeyForItem returns Key from vault.Item or nil if not recognized as a Key.
func KeyForItem(item *Item) (keys.Key, error) {
	switch item.Type {
	case string(keys.X25519):
		return x25519KeyForItem(item)
	case string(keys.X25519Public):
		return x25519PublicKeyForItem(item)
	case string(keys.EdX25519):
		return edx25519KeyForItem(item)
	case string(keys.EdX25519Public):
		return edx25519PublicKeyForItem(item)
	default:
		return nil, nil
	}
}
