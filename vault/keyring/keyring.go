package keyring

import (
	"bytes"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Keyring vault.
type Keyring struct {
	*vault.Vault
}

// New keyring vault.
func New(vlt *vault.Vault) *Keyring {
	return &Keyring{vlt}
}

// keyItemType for a generic api.Key.
const keyItemType = "key"

func newItemForKey(key *api.Key) (*vault.Item, error) {
	if key.ID == "" {
		return nil, errors.Errorf("no key id")
	}
	b, err := marshalKey(key)
	if err != nil {
		return nil, err
	}
	item := vault.NewItem(key.ID.String(), b, keyItemType, tsutil.ParseMillis(key.CreatedAt))
	return item, nil
}

func marshalKey(key *api.Key) ([]byte, error) {
	return msgpack.Marshal(key)
}

// Get key from vault.
func (v *Keyring) Get(id keys.ID) (*api.Key, error) {
	item, err := v.Vault.Get(id.String())
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return keyForItem(item)
}

// Save key to vault.
func (v *Keyring) Save(key *api.Key) error {
	if key == nil {
		return errors.Errorf("nil key")
	}

	if key.ID == "" {
		return errors.Errorf("no key id")
	}

	item, err := newItemForKey(key)
	if err != nil {
		return err
	}
	if err := v.Set(item); err != nil {
		return err
	}

	return nil
}

// Key for Item or nil if not a recognized key type.
func keyForItem(i *vault.Item) (*api.Key, error) {
	switch i.Type {
	case keyItemType:
		return unmarshalKey(i.Data)
	}
	return keyV1ForItem(i)
}

func unmarshalKey(b []byte) (*api.Key, error) {
	var key api.Key
	if err := msgpack.Unmarshal(b, &key); err != nil {
		return nil, err
	}
	return &key, nil
}

// List keys from the vault.
func (v *Keyring) List() ([]*api.Key, error) {
	items, err := v.Items()
	if err != nil {
		return nil, err
	}
	out := make([]*api.Key, 0, len(items))
	for _, i := range items {
		key, err := keyForItem(i)
		if err != nil {
			// Skip keys that don't resolve, which could happen if older clients
			// load newer keys.
			// logger.Errorf("Failed to resolve key (%s): %v", i.ID, err)
			continue
		}
		if key == nil {
			continue
		}
		out = append(out, key)
	}
	return out, nil
}

// ImportKey imports key into the vault.
func (v *Keyring) ImportKey(msg string, password string) (*api.Key, error) {
	key, err := api.DecodeKey(msg, password)
	if err != nil {
		return nil, err
	}
	if err := v.Save(key); err != nil {
		return nil, err
	}
	return key, nil
}

// ExportKey a Key from the vault.
func (v *Keyring) ExportKey(id keys.ID, password string) (string, error) {
	item, err := v.Vault.Get(id.String())
	if err != nil {
		return "", err
	}
	if item == nil {
		return "", keys.NewErrNotFound(id.String())
	}
	key, err := keyForItem(item)
	if err != nil {
		return "", err
	}
	return api.EncodeKey(key, password)
}

// EdX25519Keys implements wormhole.Keyring.
func (v *Keyring) EdX25519Keys() ([]*keys.EdX25519Key, error) {
	ks, err := v.List()
	if err != nil {
		return nil, err
	}
	out := make([]*keys.EdX25519Key, 0, len(ks))
	for _, key := range ks {
		sk := key.AsEdX25519()
		if sk == nil {
			continue
		}
		out = append(out, sk)
	}
	return out, nil
}

// X25519Keys implements saltpack.Keyring.
func (v *Keyring) X25519Keys() ([]*keys.X25519Key, error) {
	ks, err := v.List()
	if err != nil {
		return nil, err
	}
	out := make([]*keys.X25519Key, 0, len(ks))
	for _, key := range ks {
		bk := key.AsX25519()
		if bk == nil {
			continue
		}
		out = append(out, bk)
	}
	return out, nil
}

// FindEdX25519PublicKey searches all our EdX25519 public keys for a match to a converted
// X25519 public key.
func (v *Keyring) FindEdX25519PublicKey(kid keys.ID) (*keys.EdX25519PublicKey, error) {
	// logger.Debugf("Finding edx25519 key from an x25519 key %s", kid)
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
			// logger.Debugf("Found ed25519 key %s", spk.ID())
			return spk, nil
		}
	}
	// logger.Debugf("EdX25519 public key not found (for X25519 public key)")
	return nil, err
}

// EdX25519PublicKeys from the vault.
// Includes public keys of EdX25519Key's.
func (v *Keyring) EdX25519PublicKeys() ([]*keys.EdX25519PublicKey, error) {
	ks, err := v.List()
	if err != nil {
		return nil, err
	}
	out := make([]*keys.EdX25519PublicKey, 0, len(ks))
	for _, key := range ks {
		pk := key.AsEdX25519Public()
		if pk == nil {
			continue
		}
		out = append(out, pk)
	}
	return out, nil
}

// EdX25519Key ...
func (v *Keyring) EdX25519Key(kid keys.ID) (*keys.EdX25519Key, error) {
	key, err := v.Get(kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, nil
	}
	return key.AsEdX25519(), nil
}
