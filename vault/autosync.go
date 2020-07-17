package vault

import (
	"context"
	"time"
)

// AutoSync performs sync unless disabled.
func (v *Vault) AutoSync(ctx context.Context, stale time.Duration) (bool, error) {
	// If auto sync disabled, skip...
	disabled, err := v.autoSyncDisabled()
	if err != nil {
		return false, err
	}
	if disabled {
		logger.Debugf("Auto sync disabled")
		return false, nil
	}
	// If we've never synced before, skip...
	last, err := v.lastSync()
	if err != nil {
		return false, err
	}
	if last.IsZero() {
		logger.Debugf("Never synced")
		return false, nil
	}

	diff := v.clock().Sub(last)
	if diff >= 0 && diff < stale {
		logger.Debugf("Already synced recently")
		return false, nil
	}

	logger.Debugf("Last synced: %s", last)
	if err := v.Sync(ctx); err != nil {
		return true, err
	}

	return true, nil
}

// CheckAutoSync will attempt sync unless disabled and if we haven't synced
// in a stale duration.
// The afterFn always runs no matter what.
func (v *Vault) CheckAutoSync(stale time.Duration, afterFn func()) {
	go func() {
		synced, err := v.AutoSync(context.Background(), stale)
		if err != nil {
			logger.Errorf("Failed to auto sync: %v", err)
		}
		if synced {
			logger.Infof("Synced (auto)")
		}
		if afterFn != nil {
			afterFn()
		}
	}()
}
