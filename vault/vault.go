package vault

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// Vault stores keys and secrets.
type Vault struct {
	store  Store
	remote *client.Client
	clock  func() time.Time

	mk *[32]byte
	rk *keys.EdX25519Key

	inc    int
	incMax int

	protocol protocol
	subs     *subscribers
}

// New vault.
func New(st Store, opt ...Option) *Vault {
	opts := newOptions(opt...)
	return &Vault{
		store:    st,
		clock:    opts.Clock,
		protocol: opts.protocol,
		subs:     newSubscribers(),
	}
}

// Store for vault.
func (v *Vault) Store() Store {
	return v.store
}

// SetRemote sets the remote.
func (v *Vault) SetRemote(remote *client.Client) {
	v.remote = remote
}

// SetRemoteKey sets the remote key.
func (v *Vault) SetRemoteKey(rk *keys.EdX25519Key) {
	v.rk = rk
}

// SetMasterKey sets the master key.
func (v *Vault) SetMasterKey(mk *[32]byte) {
	v.mk = mk
}

// MasterKey returns master key, if unlocked.
// It's not recommended to use this key for anything other than possibly
// deriving new keys.
func (v *Vault) MasterKey() *[32]byte {
	return v.mk
}

// Set vault item.
func (v *Vault) Set(item *Item) error {
	b, err := encryptItem(item, v.mk)
	if err != nil {
		return err
	}
	path := v.protocol.Path(itemEntity, item.ID)
	if err := v.store.Set(path, b); err != nil {
		return err
	}
	return v.addPending(item.ID, b)
}

func (v *Vault) addPending(id string, b []byte) error {
	cpath := v.protocol.Path(configEntity, "increment")
	inc, err := v.Increment(cpath)
	if err != nil {
		return err
	}
	path := v.protocol.Path(pendingEntity, id, inc)
	if err := v.store.Set(path, b); err != nil {
		return err
	}
	return nil
}

// Get vault item.
func (v *Vault) Get(id string) (*Item, error) {
	path := v.protocol.Path(itemEntity, id)
	b, err := v.store.Get(path)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	item, err := decryptItem(b, v.mk)
	if err != nil {
		return nil, err
	}
	if item.ID != id {
		return nil, errors.Errorf("item id mismatch %s != %s", item.ID, id)
	}
	return item, nil
}

// Delete vault item.
func (v *Vault) Delete(id string) (bool, error) {
	exists, err := v.Exists(id)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	// Delete sets with empty bytes.
	item := &Item{
		ID:   id,
		Data: []byte{},
	}
	b, err := encryptItem(item, v.mk)
	if err != nil {
		return false, err
	}

	path := v.protocol.Path(itemEntity, id)
	if _, err := v.store.Delete(path); err != nil {
		return false, err
	}

	if err := v.addPending(item.ID, b); err != nil {
		return false, err
	}
	return true, nil
}

// Exists returns true if item exists.
func (v *Vault) Exists(id string) (bool, error) {
	path := v.protocol.Path(itemEntity, id)
	return v.store.Exists(path)
}

// Sync vault.
func (v *Vault) Sync(ctx context.Context) error {
	if err := v.push(ctx); err != nil {
		return errors.Wrapf(err, "failed to push vault")
	}
	if err := v.pull(ctx); err != nil {
		return errors.Wrapf(err, "failed to pull vault")
	}
	return nil
}

// pending returns list of pending items awaiting push.
func (v *Vault) pending(id string) ([]*Item, []string, error) {
	path := v.protocol.Path(pendingEntity, id)
	iter, err := v.store.Documents(ds.Prefix(path))
	if err != nil {
		return nil, nil, err
	}
	defer iter.Release()
	items := []*Item{}
	paths := []string{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, nil, err
		}
		if doc == nil {
			break
		}
		item, err := decryptItem(doc.Data, v.mk)
		if err != nil {
			return nil, nil, err
		}
		paths = append(paths, doc.Path)
		items = append(items, item)
	}
	return items, paths, nil
}

// Items to list.
func (v *Vault) Items() ([]*Item, error) {
	path := v.protocol.Path(itemEntity, "")
	iter, err := v.store.Documents(ds.Prefix(path))
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	items := []*Item{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if doc == nil {
			break
		}
		// Skip protocol v1 reserved entries.
		if strings.HasPrefix(doc.Path, "#") {
			continue
		}
		item, err := decryptItem(doc.Data, v.mk)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// History returns history of an item.
// Items with empty data signify deleted items.
func (v *Vault) History(id string) ([]*Item, error) {
	path := v.protocol.Path(historyEntity, id)
	iter, err := v.store.Documents(ds.Prefix(path))
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	items := []*Item{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if doc == nil {
			break
		}
		item, err := decryptItem(doc.Data, v.mk)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	pending, _, err := v.pending(id)
	if err != nil {
		return nil, err
	}
	items = append(items, pending...)

	return items, nil
}

func (v *Vault) push(ctx context.Context) error {
	if v.remote == nil {
		return errors.Errorf("no vault remote set")
	}
	if v.rk == nil {
		return errors.Errorf("no remote key")
	}

	paths := []string{}
	items := []*client.VaultItem{}

	pending, pp, err := v.pending("")
	if err != nil {
		return err
	}
	for _, item := range pending {
		b, err := encryptItem(item, v.mk)
		if err != nil {
			return err
		}
		items = append(items, &client.VaultItem{Data: b})
	}

	if len(items) > 0 {
		logger.Infof("Pushing %d vault items", len(items))
		if err := v.remote.VaultSave(ctx, v.rk, items); err != nil {
			return err
		}
		logger.Infof("Removing %d pending vault items", len(paths))
		if err := deleteAll(v.store, pp); err != nil {
			return err
		}
	}

	return nil
}

// pull changes from remote.
func (v *Vault) pull(ctx context.Context) error {
	if v.mk == nil {
		return ErrLocked
	}
	if v.remote == nil {
		return errors.Errorf("no vault remote set")
	}
	if v.rk == nil {
		return errors.Errorf("no remote key")
	}

	cpath := v.protocol.Path(configEntity, "version")
	b, err := v.store.Get(cpath)
	if err != nil {
		return err
	}
	version := string(b)

	logger.Infof("Getting vault items")
	vault, err := v.remote.Vault(ctx, v.rk, client.VaultVersion(version))
	if err != nil {
		return err
	}

	if vault != nil {
		for _, vaultItem := range vault.Items {
			item, err := decryptItem(vaultItem.Data, v.mk)
			if err != nil {
				return err
			}
			item.Timestamp = vaultItem.Timestamp
			b, err := encryptItem(item, v.mk)
			if err != nil {
				return err
			}

			// TODO: Is nanosecond available from server? Use it?
			ts := fmt.Sprintf("%015d", tsutil.Millis(item.Timestamp))
			hpath := v.protocol.Path(historyEntity, item.ID, ts)
			if err := v.store.Set(hpath, b); err != nil {
				return err
			}
			ipath := v.protocol.Path(itemEntity, item.ID)
			if len(item.Data) == 0 {
				if _, err := v.store.Delete(ipath); err != nil {
					return err
				}
			} else {
				if err := v.store.Set(ipath, b); err != nil {
					return err
				}
			}
		}

		// Update version
		if err := v.store.Set(cpath, []byte(vault.Version)); err != nil {
			return err
		}
	}

	return nil
}

// Increment returns the current increment as an orderable string that persists
// across opens.
// => 000000000000001, 000000000000002 ...
// This is batched. When the increment runs out for the current batch, it
// gets a new batch.
// The increment value is saved in the database at the specified path.
// There may be large gaps between increments (of batch size) after re-opens.
func (v *Vault) Increment(path string) (string, error) {
	if v.inc == 0 || v.inc >= v.incMax {
		if err := v.increment(path); err != nil {
			return "", err
		}
	}
	v.inc++
	if v.inc > 999999999999999 {
		panic("index too large")
	}
	return fmt.Sprintf("%015d", v.inc), nil
}

const incrementBatchSize = 1000

// Increment value from path.
func (v *Vault) increment(path string) error {
	b, err := v.store.Get(path)
	if err != nil {
		return err
	}

	inc := 0
	if b != nil {
		i, err := strconv.Atoi(string(b))
		if err != nil {
			return err
		}
		inc = i
	}

	if err := v.store.Set(path, []byte(strconv.Itoa(inc+incrementBatchSize))); err != nil {
		return err
	}

	logger.Debugf("Setting increment batch: %d", inc)
	v.inc = inc
	v.incMax = inc + incrementBatchSize - 1
	return nil
}

// Spew to out.
func (v *Vault) Spew(prefix string, out io.Writer) error {
	iter, err := v.store.Documents(ds.Prefix(prefix))
	if err != nil {
		return err
	}
	defer iter.Release()
	if err := ds.SpewOut(iter, out); err != nil {
		return err
	}
	return nil
}
