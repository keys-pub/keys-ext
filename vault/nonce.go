package vault

import (
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
)

func (v *Vault) checkNonce(n string) error {
	b, err := v.store.Get(ds.Path("db", "nonce", n))
	if err != nil {
		return err
	}
	if b != nil {
		return errors.Errorf("nonce collision %s", n)
	}
	return nil
}

func (v *Vault) commitNonce(n string) error {
	if err := v.store.Set(ds.Path("db", "nonce", n), []byte{0x01}); err != nil {
		return err
	}
	return nil
}
