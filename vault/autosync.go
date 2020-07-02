package vault

import (
	"context"
)

func (v *Vault) autoSync(ctx context.Context) (bool, error) {
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
	logger.Debugf("Last synced: %s", last)
	if last.IsZero() {
		return false, nil
	}
	if err := v.Sync(ctx); err != nil {
		return true, err
	}

	return true, nil
}

// AutoSync will attempt sync (unless disabled or sync has never run).
// The afterFn always runs no matter what.
func (v *Vault) AutoSync(afterFn func()) {
	go func() {
		synced, err := v.autoSync(context.Background())
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
