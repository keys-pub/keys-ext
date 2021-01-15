package keyring

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/api"
	"github.com/pkg/errors"
)

// Keys used to be stored as item data directly instead of as a marshaled
// api.Key.
func keyV1ForItem(i *vault.Item) (*api.Key, error) {
	switch i.Type {
	case "ed25519-public":
		if len(i.Data) != 32 {
			return nil, errors.Errorf("invalid key data (ed25519-public)")
		}
		k := keys.NewEdX25519PublicKey(keys.Bytes32(i.Data))
		out := api.NewKey(k)
		return out, nil
	case "edx25519":
		if len(i.Data) != 64 {
			return nil, errors.Errorf("invalid key data (edx25519)")
		}
		k := keys.NewEdX25519KeyFromPrivateKey(keys.Bytes64(i.Data))
		out := api.NewKey(k)
		return out, nil
	case "x25519-public":
		if len(i.Data) != 32 {
			return nil, errors.Errorf("invalid key data (x25519-public)")
		}
		k := keys.NewEdX25519PublicKey(keys.Bytes32(i.Data))
		out := api.NewKey(k)
		return out, nil
	case "x25519":
		if len(i.Data) != 64 {
			return nil, errors.Errorf("invalid key data (x25519)")
		}
		k := keys.NewEdX25519KeyFromPrivateKey(keys.Bytes64(i.Data))
		out := api.NewKey(k)
		return out, nil
	default:
		return nil, nil
	}
}
