package vault

import (
	"bytes"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

const keyItemType = "key"

// Key for vault.
type Key struct {
	ID   keys.ID `json:"id" msgpack:"id,omitempty"`
	Data []byte  `json:"data" msgpack:"dat,omitempty"`
	Type string  `json:"type" msgpack:"typ,omitempty"`

	Notes string `json:"notes,omitempty" msgpack:"nt,omitempty"`

	CreatedAt time.Time `json:"createdAt,omitempty" msgpack:"cts,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty" msgpack:"uts,omitempty"`
}

// NewKey from keys.Key interface.
func NewKey(k keys.Key, createdAt time.Time) *Key {
	return &Key{
		ID:        k.ID(),
		Data:      k.Bytes(),
		Type:      string(k.Type()),
		CreatedAt: createdAt,
	}
}

func newItemForKey(key *Key) (*Item, error) {
	if key.ID == "" {
		return nil, errors.Errorf("no secret id")
	}
	b, err := marshalKey(key)
	if err != nil {
		return nil, err
	}
	item := NewItem(key.ID.String(), b, keyItemType, key.CreatedAt)
	return item, nil
}

func marshalKey(key *Key) ([]byte, error) {
	return msgpack.Marshal(key)
}

// Key from vault.
func (v *Vault) Key(id keys.ID) (*Key, error) {
	item, err := v.Get(id.String())
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return item.Key()
}

// SaveKey saves key to vault.
func (v *Vault) SaveKey(key *Key) (*Key, bool, error) {
	if key == nil {
		return nil, false, errors.Errorf("nil secret")
	}

	if key.ID == "" {
		return nil, false, errors.Errorf("no key id")
	}

	item, err := v.Get(key.ID.String())
	if err != nil {
		return nil, false, err
	}

	updated := false
	if item != nil {
		key.UpdatedAt = v.Now()
		b, err := marshalKey(key)
		if err != nil {
			return nil, false, err
		}
		item.Data = b
		if err := v.Set(item); err != nil {
			return nil, false, err
		}
		updated = true
	} else {
		now := v.Now()
		key.CreatedAt = now
		key.UpdatedAt = now

		item, err := newItemForKey(key)
		if err != nil {
			return nil, false, err
		}
		if err := v.Set(item); err != nil {
			return nil, false, err
		}
	}

	return key, updated, nil
}

// Key for Item or nil if not a recognized key type.
func (i *Item) Key() (*Key, error) {
	switch i.Type {
	case keyItemType:
		return unmarshalKey(i.Data)
	// Keys used to be stored as item data directly instead of as a marshaled vault.Key.
	case string(keys.X25519), string(keys.X25519Public), string(keys.EdX25519), string(keys.EdX25519Public):
		return &Key{
			ID:        keys.ID(i.ID),
			Data:      i.Data,
			Type:      i.Type,
			CreatedAt: i.CreatedAt,
			UpdatedAt: i.CreatedAt,
		}, nil
	default:
		return nil, nil
	}
}

func unmarshalKey(b []byte) (*Key, error) {
	var key Key
	if err := msgpack.Unmarshal(b, &key); err != nil {
		return nil, err
	}
	return &key, nil
}

// Keys returns keys from the vault.
func (v *Vault) Keys() ([]*Key, error) {
	items, err := v.Items()
	if err != nil {
		return nil, err
	}
	out := make([]*Key, 0, len(items))
	for _, i := range items {
		key, err := i.Key()
		if err != nil {
			return nil, err
		}
		if key == nil {
			continue
		}
		out = append(out, key)
	}
	return out, nil
}

// ImportSaltpack imports key into the vault from a Saltpack message.
func (v *Vault) ImportSaltpack(msg string, password string, isHTML bool) (*Key, error) {
	kkey, err := keys.DecodeSaltpackKey(msg, password, isHTML)
	if err != nil {
		return nil, err
	}
	key := NewKey(kkey, v.Now())
	out, _, err := v.SaveKey(key)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExportSaltpack exports a Key from the vault to a Saltpack message.
func (v *Vault) ExportSaltpack(id keys.ID, password string) (string, error) {
	item, err := v.Get(id.String())
	if err != nil {
		return "", err
	}
	if item == nil {
		return "", keys.NewErrNotFound(id.String())
	}
	key, err := item.Key()
	if err != nil {
		return "", err
	}
	var brand keys.Brand
	switch key.Type {
	case string(keys.EdX25519):
		brand = keys.EdX25519Brand
	case string(keys.X25519):
		brand = keys.X25519Brand
	default:
		return "", errors.Errorf("failed to encode to saltpack: unsupported key %s", key.Type)
	}
	out := keys.EncryptWithPassword(key.Data, password)
	return encoding.EncodeSaltpack(out, string(brand)), nil
}

// EdX25519Keys implements wormhole.Keyring.
func (v *Vault) EdX25519Keys() ([]*keys.EdX25519Key, error) {
	ks, err := v.Keys()
	if err != nil {
		return nil, err
	}
	out := make([]*keys.EdX25519Key, 0, len(ks))
	for _, key := range ks {
		switch key.Type {
		case string(keys.EdX25519):
			sk, err := key.AsEdX25519()
			if err != nil {
				return nil, err
			}
			out = append(out, sk)
		}
	}
	return out, nil
}

// X25519Keys implements saltpack.Keyring.
func (v *Vault) X25519Keys() ([]*keys.X25519Key, error) {
	ks, err := v.Keys()
	if err != nil {
		return nil, err
	}
	out := make([]*keys.X25519Key, 0, len(ks))
	for _, key := range ks {
		switch key.Type {
		case string(keys.X25519), string(keys.EdX25519):
			bk, err := key.AsX25519()
			if err != nil {
				return nil, err
			}
			out = append(out, bk)
		}
	}
	return out, nil
}

// FindEdX25519PublicKey searches all our EdX25519 public keys for a match to a converted
// X25519 public key.
func (v *Vault) FindEdX25519PublicKey(kid keys.ID) (*keys.EdX25519PublicKey, error) {
	logger.Debugf("Finding edx25519 key from an x25519 key %s", kid)
	if !kid.IsX25519() {
		return nil, errors.Errorf("not an x25519 key")
	}
	bpk, err := keys.NewX25519PublicKeyFromID(kid)
	if err != nil {
		return nil, err
	}
	spks, err := v.EdX25519PublicKeys()
	if err != nil {
		return nil, err
	}
	for _, spk := range spks {
		if bytes.Equal(spk.X25519PublicKey().Bytes(), bpk.Bytes()) {
			logger.Debugf("Found ed25519 key %s", spk.ID())
			return spk, nil
		}
	}
	logger.Debugf("EdX25519 public key not found (for X25519 public key)")
	return nil, err
}

// EdX25519PublicKeys from the vault.
// Includes public keys of EdX25519Key's.
func (v *Vault) EdX25519PublicKeys() ([]*keys.EdX25519PublicKey, error) {
	ks, err := v.Keys()
	if err != nil {
		return nil, err
	}
	out := make([]*keys.EdX25519PublicKey, 0, len(ks))
	for _, key := range ks {
		pk, err := key.AsEdX25519Public()
		if err != nil {
			return nil, err
		}
		if pk == nil {
			continue
		}
		out = append(out, pk)
	}
	return out, nil
}

// EdX25519Key ...
func (v *Vault) EdX25519Key(kid keys.ID) (*keys.EdX25519Key, error) {
	key, err := v.Key(kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, nil
	}
	return key.AsEdX25519()
}

// AsEdX25519 returns a *EdX25519Key.
func (k *Key) AsEdX25519() (*keys.EdX25519Key, error) {
	if k.Type != string(keys.EdX25519) {
		return nil, errors.Errorf("type %s != %s", k.Type, keys.EdX25519)
	}
	b := k.Data
	if len(b) != 64 {
		return nil, errors.Errorf("invalid number of bytes for ed25519 private key")
	}
	out := keys.NewEdX25519KeyFromPrivateKey(keys.Bytes64(b))
	return out, nil
}

// AsX25519 returns a X25519Key.
// If key is a EdX25519Key, it's converted to a X25519Key.
func (k *Key) AsX25519() (*keys.X25519Key, error) {
	switch k.Type {
	case string(keys.X25519):
		bk := keys.NewX25519KeyFromPrivateKey(keys.Bytes32(k.Data))
		return bk, nil
	case string(keys.EdX25519):
		sk, err := k.AsEdX25519()
		if err != nil {
			return nil, err
		}
		return sk.X25519Key(), nil
	default:
		return nil, errors.Errorf("type %s != %s (or %s)", k.Type, keys.X25519, keys.EdX25519)
	}
}

// AsEdX25519Public returns a *EdX25519PublicKey.
func (k *Key) AsEdX25519Public() (*keys.EdX25519PublicKey, error) {
	switch k.Type {
	case string(keys.EdX25519):
		sk, err := k.AsEdX25519()
		if err != nil {
			return nil, err
		}
		return sk.PublicKey(), nil
	case string(keys.EdX25519Public):
		b := k.Data
		if len(b) != 32 {
			return nil, errors.Errorf("invalid number of bytes for ed25519 public key")
		}
		out := keys.NewEdX25519PublicKey(keys.Bytes32(b))
		return out, nil
	default:
		return nil, errors.Errorf("type %s != %s (or %s)", k.Type, keys.EdX25519Public, keys.EdX25519)
	}
}

// AsX25519Public returns a X25519PublicKey.
func (k *Key) AsX25519Public() (*keys.X25519PublicKey, error) {
	switch k.Type {
	case string(keys.X25519):
		sk, err := k.AsX25519()
		if err != nil {
			return nil, err
		}
		return sk.PublicKey(), nil
	case string(keys.X25519Public):
		b := k.Data
		if len(b) != 32 {
			return nil, errors.Errorf("invalid number of bytes for x25519 public key")
		}
		out := keys.NewX25519PublicKey(keys.Bytes32(b))
		return out, nil
	default:
		return nil, errors.Errorf("type %s != %s (or %s)", k.Type, keys.X25519Public, keys.X25519)
	}
}
