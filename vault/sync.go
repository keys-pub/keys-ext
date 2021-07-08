package vault

import (
	"context"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// SyncStatus is status of sync.
type SyncStatus struct {
	KID      keys.ID
	Salt     []byte
	SyncedAt time.Time
}

// Sync vault.
func (v *Vault) Sync(ctx context.Context) error {
	v.mtx.Lock()
	defer v.mtx.Unlock()
	logger.Infof("Syncing...")

	// What happens on connection failures, context cancellation?
	//
	// If we fail during push, we could push duplicates on the next push.
	// But the duplicates would only show up in the history.
	// We could de-dupe on the clients, but this is probably rare enough to
	// ignore for now (TODO: resolve possible duplicate push events).
	//
	// If we fail after pull, we could pull duplicates, but this is ok since
	// the partial data would be overwritten on the next pull.

	if err := v.push(ctx); err != nil {
		return errors.Wrapf(err, "failed to push vault")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := v.pull(ctx); err != nil {
		return errors.Wrapf(err, "failed to pull vault")
	}

	if err := v.setLastSync(time.Now()); err != nil {
		return err
	}

	return nil
}

// SyncStatus returns status for sync, or nil, if no sync has been performed.
func (v *Vault) SyncStatus() (*SyncStatus, error) {
	lastSync, err := v.lastSync()
	if err != nil {
		return nil, err
	}
	if lastSync.IsZero() {
		return nil, nil
	}
	remote := v.Remote()
	if remote == nil {
		return nil, nil
	}
	return &SyncStatus{
		KID:      remote.Key.ID(),
		Salt:     remote.Salt,
		SyncedAt: lastSync,
	}, nil
}

// Unsync removes vault from the remote and resets the vault log.
//
// The steps for "unsyncing" are:
// - Delete the vault from the server
// - Reset log (move pull into push)
// - Clear status (last synced, push, pull, nonces, rsalt)
// - Clear remote
func (v *Vault) Unsync(ctx context.Context) error {
	v.mtx.Lock()
	defer v.mtx.Unlock()
	logger.Infof("Unsyncing...")

	if v.remote == nil {
		return errors.Errorf("no remote set")
	}
	if v.mk == nil {
		return errors.Errorf("vault is locked")
	}

	// Delete vault from the server
	if err := v.client.VaultDelete(ctx, v.remote.Key); err != nil {
		return err
	}

	// Reset log (move pull into push)
	if err := v.resetLog(); err != nil {
		return err
	}

	// Clear status (last synced,index, nonces)
	if err := v.setLastSync(time.Time{}); err != nil {
		return err
	}
	if err := v.setPullIndex(0); err != nil {
		return err
	}

	// Clear remote
	if err := v.clearRemote(); err != nil {
		return err
	}

	return nil
}

func (v *Vault) resetLog() error {
	push, err := v.store.List(&ListOptions{Prefix: dstore.Path("push")})
	if err != nil {
		return err
	}

	pull, err := v.store.List(&ListOptions{Prefix: dstore.Path("pull")})
	if err != nil {
		return err
	}
	if len(pull) == 0 {
		return nil
	}

	if err := v.setPushIndex(int64(len(pull) + len(push))); err != nil {
		return err
	}

	// Move push to the end
	index := int64(len(pull))
	for _, doc := range push {
		index++
		path := dstore.PathFrom(doc.Path, 2)
		push := dstore.Path("push", pad(index), path)
		if err := v.store.Set(push, doc.Data); err != nil {
			return err
		}
	}

	// Move pull back to push
	index = int64(0)
	for _, doc := range pull {
		index++
		var event api.Event
		if err := msgpack.Unmarshal(doc.Data, &event); err != nil {
			return err
		}
		path := dstore.PathFrom(doc.Path, 2)
		push := dstore.Path("push", pad(index), path)
		if err := v.store.Set(push, event.Data); err != nil {
			return err
		}
		if _, err := v.store.Delete(doc.Path); err != nil {
			return err
		}
	}

	return nil
}

// SyncEnabled returns true if sync is enabled.
// Sync is enabled by performing a sync and not having sync disabled.
func (v *Vault) SyncEnabled() (bool, error) {
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

	return true, nil
}

func (v *Vault) shouldCheck(expire time.Duration) (bool, error) {
	v.checkMtx.Lock()
	defer v.checkMtx.Unlock()

	enabled, err := v.SyncEnabled()
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}

	diffCheck := v.clock.Now().Sub(v.checkedAt)
	if diffCheck >= 0 && diffCheck < expire {
		logger.Debugf("Already checked recently")
		return false, nil
	}
	v.checkedAt = v.clock.Now()

	last, err := v.lastSync()
	if err != nil {
		return false, err
	}
	logger.Debugf("Last synced: %s", last)
	diffLast := v.clock.Now().Sub(last)
	if diffLast >= 0 && diffLast < expire {
		logger.Debugf("Already synced recently")
		return false, nil
	}

	return true, nil
}

// CheckSync performs sync unless disabled or already synced recently (within expire duration).
func (v *Vault) CheckSync(ctx context.Context, expire time.Duration) (bool, error) {
	enabled, err := v.shouldCheck(expire)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}

	if err := v.Sync(ctx); err != nil {
		return true, err
	}

	return true, nil
}

func (v *Vault) pullIndex() (int64, error) {
	return v.getInt64("/sync/pull")
}

func (v *Vault) setPullIndex(n int64) error {
	return v.setInt64("/sync/pull", n)
}

func (v *Vault) pushIndex() (int64, error) {
	return v.getInt64("/sync/push")
}

func (v *Vault) setPushIndex(n int64) error {
	return v.setInt64("/sync/push", n)
}

func (v *Vault) pushIndexNext() (int64, error) {
	n, err := v.pushIndex()
	if err != nil {
		return 0, err
	}
	n++
	if err := v.setPushIndex(n); err != nil {
		return 0, err
	}
	return n, nil
}

func (v *Vault) autoSyncDisabled() (bool, error) {
	return v.getBool("/sync/autoDisabled")
}

// func (v *Vault) setAutoSyncDisabled(b bool) error {
// 	return v.setBool("/sync/autoDisabled", b)
// }

func (v *Vault) lastSync() (time.Time, error) {
	return v.getTime("/sync/lastSync")
}

func (v *Vault) setLastSync(tm time.Time) error {
	return v.setTime("/sync/lastSync", tm)
}

func (v *Vault) setRemoteSalt(b []byte) error {
	return v.setValue("/sync/rsalt", b)
}

func (v *Vault) getRemoteSalt(init bool) ([]byte, error) {
	salt, err := v.getValue("/sync/rsalt")
	if err != nil {
		return nil, err
	}
	if salt == nil && init {
		salt = keys.RandBytes(32)
		if err := v.setRemoteSalt(salt); err != nil {
			return nil, err
		}
	}
	return salt, nil
}
