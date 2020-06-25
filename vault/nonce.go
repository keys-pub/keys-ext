package vault

import (
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

func (v *Vault) checkNonce(n []byte) error {
	nb := encoding.MustEncode(n, encoding.Base62)
	b, err := v.store.Get(ds.Path("db", "nonce", nb))
	if err != nil {
		return err
	}
	if b != nil {
		return errors.Errorf("nonce collision %s", nb)
	}
	return nil
}

func (v *Vault) commitNonce(n []byte) error {
	nb := encoding.MustEncode(n, encoding.Base62)
	if err := v.store.Set(ds.Path("db", "nonce", nb), []byte{0x01}); err != nil {
		return err
	}
	return nil
}
