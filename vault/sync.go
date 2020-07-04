package vault

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// Sync vault.
func (v *Vault) Sync(ctx context.Context) error {
	v.mtx.Lock()
	defer v.mtx.Unlock()
	logger.Infof("Syncing...")

	if err := v.push(ctx); err != nil {
		return errors.Wrapf(err, "failed to push vault (sync)")
	}
	if err := v.pull(ctx); err != nil {
		return errors.Wrapf(err, "failed to pull vault (sync)")
	}

	if err := v.setLastSync(time.Now()); err != nil {
		return err
	}

	return nil
}

// Unsync removes vault from remote.
func (v *Vault) Unsync(ctx context.Context) error {
	if err := v.remote.VaultDelete(ctx, v.rk); err != nil {
		return err
	}
	return nil
}
