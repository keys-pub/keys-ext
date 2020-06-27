package vault

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// InitRemote creates Vault from remote.
func (v *Vault) InitRemote(ctx context.Context, rk *keys.EdX25519Key) error {
	if v.remote == nil {
		return errors.Errorf("no vault remote set")
	}
	if rk == nil {
		return errors.Errorf("no remote key")
	}

	empty, err := v.IsEmpty()
	if err != nil {
		return err
	}
	if !empty {
		return errors.Errorf("vault not empty, can only be initialized from remote if empty")
	}

	logger.Infof("Requesting remote vault...")
	vault, err := v.remote.Vault(ctx, rk)
	if err != nil {
		return err
	}

	if err := v.saveRemoveVault(vault); err != nil {
		return err
	}

	// Setting remote key, which unlock will check to make sure it's not
	// different from what is expected.
	v.rk = rk
	return nil
}
