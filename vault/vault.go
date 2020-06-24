package vault

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
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

	inc    int64
	incMax int64
	subs   *subscribers
}

// New vault.
func New(st Store, opt ...Option) *Vault {
	opts := newOptions(opt...)
	return &Vault{
		store: st,
		clock: opts.Clock,
		subs:  newSubscribers(),
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
	return v.set(ds.Path("item", item.ID), b)
}

func (v *Vault) set(path string, b []byte) error {
	if err := v.store.Set(path, b); err != nil {
		return err
	}
	return v.addToPush(path, b)
}

func (v *Vault) addToPush(path string, b []byte) error {
	inc, err := v.Increment(1)
	if err != nil {
		return err
	}
	ppath := ds.Path("push", inc, path)
	if err := v.store.Set(ppath, b); err != nil {
		return err
	}
	return nil
}

// Get vault item.
func (v *Vault) Get(id string) (*Item, error) {
	path := ds.Path("item", id)
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
	if len(item.Data) == 0 {
		return nil, nil
	}
	return item, nil
}

// Delete vault item.
func (v *Vault) Delete(id string) (bool, error) {
	item, err := v.Get(id)
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, nil
	}

	// Delete clears bytes
	item.Data = nil
	b, err := encryptItem(item, v.mk)
	if err != nil {
		return false, err
	}
	if err := v.set(ds.Path("item", item.ID), b); err != nil {
		return false, err
	}

	return true, nil
}

// Sync vault.
func (v *Vault) Sync(ctx context.Context) error {
	if err := v.push(ctx); err != nil {
		return errors.Wrapf(err, "failed to push vault (sync)")
	}
	if err := v.Pull(ctx); err != nil {
		return errors.Wrapf(err, "failed to pull vault (sync)")
	}
	return nil
}

// Items to list.
func (v *Vault) Items() ([]*Item, error) {
	path := ds.Path("item")
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
	changes := []*client.VaultChange{}

	// Get changes from push.
	path := ds.Path("push")
	iter, err := v.store.Documents(ds.Prefix(path))
	if err != nil {
		return err
	}
	defer iter.Release()
	for {
		doc, err := iter.Next()
		if err != nil {
			return err
		}
		if doc == nil {
			break
		}
		logger.Debugf("Push %s", doc.Path)
		paths = append(paths, doc.Path)
		path := ds.PathFrom(doc.Path, 2)
		change := &client.VaultChange{Path: path, Data: doc.Data, Nonce: api.GenerateNonce()}
		changes = append(changes, change)
	}

	if len(changes) > 0 {
		logger.Infof("Pushing %d vault changes", len(changes))
		if err := v.remote.VaultChanged(ctx, v.rk, changes); err != nil {
			return err
		}
		logger.Infof("Removing %d from push", len(paths))
		if err := deleteAll(v.store, paths); err != nil {
			return err
		}
	}

	return nil
}

func (v *Vault) version() (string, error) {
	b, err := v.store.Get(ds.Path("db", "version"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Pull changes from remote.
// Does NOT require Unlock.
func (v *Vault) Pull(ctx context.Context) error {
	if v.remote == nil {
		return errors.Errorf("no vault remote set")
	}
	if v.rk == nil {
		return errors.Errorf("no remote key")
	}

	version, err := v.version()
	if err != nil {
		return err
	}

	logger.Infof("Getting vault items")
	vault, err := v.remote.Vault(ctx, v.rk, client.VaultVersion(version))
	if err != nil {
		return err
	}

	if vault != nil {
		for _, change := range vault.Changes {
			logger.Debugf("Pull %s", change.Path)
			if change.Path == "" {
				return errors.Errorf("invalid change (no path)")
			}
			if err := v.checkNonce(change.Nonce); err != nil {
				return err
			}

			if len(change.Data) == 0 {
				if _, err := v.store.Delete(change.Path); err != nil {
					return err
				}
			} else {
				if err := v.store.Set(change.Path, change.Data); err != nil {
					return err
				}
			}

			hpath := ds.Path("pull", pad(change.Version), tsutil.Millis(change.Timestamp), change.Path)
			if err := v.store.Set(hpath, change.Data); err != nil {
				return err
			}

			if err := v.commitNonce(change.Nonce); err != nil {
				return err
			}
		}

		// Update version
		if err := v.store.Set(ds.Path("db", "version"), []byte(vault.Version)); err != nil {
			return err
		}
	}

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

// Increment returns the current increment as an orderable string that persists
// across opens at /db/increment.
// => 000000000000001, 000000000000002 ...
// This is batched. When the increment runs out for the current batch, it
// gets a new batch.
// There may be large gaps between increments (of batch size) after re-opens.
func (v *Vault) Increment(n int64) (string, error) {
	if v.inc == 0 || (v.inc+n) >= v.incMax {
		inc, incMax, err := increment(v.store, ds.Path("db", "increment"), 1000)
		if err != nil {
			return "", err
		}
		v.inc, v.incMax = inc, incMax
	}
	v.inc = v.inc + n
	if v.inc > 999999999999999 {
		panic("index too large")
	}
	// logger.Debugf("Increment(%d) %d", n, v.inc)
	return pad(v.inc), nil
}

func increment(st Store, path string, size int64) (inc int64, incMax int64, reterr error) {
	b, err := st.Get(path)
	if err != nil {
		reterr = err
		return
	}

	if b != nil {
		i, err := strconv.Atoi(string(b))
		if err != nil {
			reterr = err
			return
		}
		inc = int64(i)
	}

	max := int(inc + size)
	if err := st.Set(path, []byte(strconv.Itoa(max))); err != nil {
		reterr = err
		return
	}

	incMax = inc + size - 1
	return
}

func pad(n int64) string {
	return fmt.Sprintf("%015d", n)
}

// IsEmpty returns true if vault is empty.
func (v *Vault) IsEmpty() (bool, error) {
	iter, err := v.store.Documents()
	if err != nil {
		return false, err
	}
	defer iter.Release()
	doc, err := iter.Next()
	if err != nil {
		return false, err
	}
	if doc == nil {
		return true, nil
	}
	return false, nil
}
