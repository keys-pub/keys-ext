package vault

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// keyV1
// Keys used to be stored as item data directly instead of as a marshaled
// api.Key.
func (i *Item) keyV1() (*api.Key, error) {
	switch i.Type {
	case "ed25519-public":
		if len(i.Data) != 32 {
			return nil, errors.Errorf("invalid key data")
		}
		k := keys.NewEdX25519PublicKey(keys.Bytes32(i.Data))
		out := api.NewKey(k)
		out.CreatedAt = tsutil.Millis(i.CreatedAt)
		out.UpdatedAt = tsutil.Millis(i.CreatedAt)
		return out, nil
	case "edx25519":
		if len(i.Data) != 64 {
			return nil, errors.Errorf("invalid key data")
		}
		k := keys.NewEdX25519KeyFromPrivateKey(keys.Bytes64(i.Data))
		out := api.NewKey(k)
		out.CreatedAt = tsutil.Millis(i.CreatedAt)
		out.UpdatedAt = tsutil.Millis(i.CreatedAt)
		return out, nil
	case "x25519-public":
		if len(i.Data) != 32 {
			return nil, errors.Errorf("invalid key data")
		}
		k := keys.NewEdX25519PublicKey(keys.Bytes32(i.Data))
		out := api.NewKey(k)
		out.CreatedAt = tsutil.Millis(i.CreatedAt)
		out.UpdatedAt = tsutil.Millis(i.CreatedAt)
		return out, nil
	case "x25519":
		if len(i.Data) != 64 {
			return nil, errors.Errorf("invalid key data")
		}
		k := keys.NewEdX25519KeyFromPrivateKey(keys.Bytes64(i.Data))
		out := api.NewKey(k)
		out.CreatedAt = tsutil.Millis(i.CreatedAt)
		out.UpdatedAt = tsutil.Millis(i.CreatedAt)
		return out, nil
	default:
		return nil, nil
	}
}
