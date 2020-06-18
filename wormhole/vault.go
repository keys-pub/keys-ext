package wormhole

import "github.com/keys-pub/keys"

type vault struct {
	keys []*keys.EdX25519Key
}

// NewVault creates a vault (for testing).
func NewVault(keys ...*keys.EdX25519Key) Vault {
	return &vault{keys: keys}
}

func (v *vault) EdX25519Key(id keys.ID) (*keys.EdX25519Key, error) {
	for _, k := range v.keys {
		if k.ID() == id {
			return k, nil
		}
	}
	return nil, nil
}

func (v *vault) EdX25519Keys() ([]*keys.EdX25519Key, error) {
	return v.keys, nil
}
