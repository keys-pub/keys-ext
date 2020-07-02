package vault

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Vault stores keys and secrets.
type Vault struct {
	mtx sync.Mutex

	store  Store
	remote *client.Client
	clock  func() time.Time

	mk *[32]byte
	rk *keys.EdX25519Key

	inc    int64
	incMax int64

	subs *subscribers
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

// Open vault.
func (v *Vault) Open() error {
	if err := v.store.Open(); err != nil {
		return errors.Wrapf(err, "failed to open vault")
	}
	return nil
}

// Close vault.
func (v *Vault) Close() error {
	if err := v.store.Close(); err != nil {
		return errors.Wrapf(err, "failed to close vault")
	}
	return nil
}

// SetRemote sets the remote.
func (v *Vault) SetRemote(remote *client.Client) {
	v.remote = remote
}

// setMasterKey sets the master key.
// You should be using Setup or Unlock instead, this is available for testing.
func (v *Vault) setMasterKey(mk *[32]byte) error {
	// Derive remote key
	rs := keys.Bytes32(keys.HKDFSHA256(mk[:], 32, nil, []byte("keys.pub/rk")))
	rk := keys.NewEdX25519KeyFromSeed(rs)
	if v.rk != nil {
		if !v.rk.Equal(rk) {
			return errors.Errorf("remote key is different than expected")
		}
	}

	v.mk = mk
	v.rk = rk
	return nil
}

// MasterKey returns master key, if unlocked.
// The master key is used to encrypt items in the vault.
// It's not recommended to use this key for anything other than possibly
// deriving new keys.
// TODO: Point to spec.
func (v *Vault) MasterKey() *[32]byte {
	return v.mk
}

// RemoteKey returns remote key, if unlocked.
// The remote key is used to encrypt and verify remote items.
// This encryption happens on top of encryption using the master key.
// It's not recommended to use this key for anything other initializing another
// vault from remote (InitRemote).
// TODO: Point to spec.
func (v *Vault) RemoteKey() *keys.EdX25519Key {
	return v.rk
}

// Set vault item.
func (v *Vault) Set(item *Item) error {
	return v.setItem(item, true)
}

func (v *Vault) setItem(item *Item, addToPush bool) error {
	b, err := encryptItem(item, v.mk)
	if err != nil {
		return err
	}
	path := ds.Path("item", item.ID)
	return v.set(path, b, addToPush)
}

func (v *Vault) set(path string, b []byte, addToPush bool) error {
	if err := v.store.Set(path, b); err != nil {
		return err
	}
	if addToPush {
		return v.addToPush(path, b)
	}
	return nil
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

	time.AfterFunc(time.Second*2, func() {
		v.AutoSync(nil)
	})

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
		logger.Debugf("Path not found %s", path)
		return nil, nil
	}
	item, err := decryptItem(b, v.mk)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	if item.ID != id {
		return nil, errors.Errorf("item id mismatch %s != %s", item.ID, id)
	}
	if len(item.Data) == 0 {
		return nil, nil
	}
	return item, err
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
	if err := v.setItem(item, true); err != nil {
		return false, err
	}

	return true, nil
}

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

// Items to list.
func (v *Vault) Items() ([]*Item, error) {
	path := ds.Path("item")
	docs, err := v.store.Documents(ds.Prefix(path))
	if err != nil {
		return nil, err
	}
	items := []*Item{}
	for _, doc := range docs {
		item, err := decryptItem(doc.Data, v.mk)
		if err != nil {
			return nil, err
		}
		if len(item.Data) == 0 {
			// TODO: Deleted item (clean it up by removing?)
			continue
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
	events := []*client.Event{}

	// Get events from push.
	path := ds.Path("push")
	docs, err := v.store.Documents(ds.Prefix(path))
	if err != nil {
		return err
	}
	var prev *client.Event
	for _, doc := range docs {
		logger.Debugf("Push %s", doc.Path)
		paths = append(paths, doc.Path)
		path := ds.PathFrom(doc.Path, 2)
		event := client.NewEvent(path, doc.Data, prev)
		events = append(events, event)
		prev = event
	}

	if len(events) > 0 {
		logger.Infof("Pushing %d vault events", len(events))
		if err := v.remote.VaultSend(ctx, v.rk, events); err != nil {
			return err
		}
		logger.Infof("Removing %d from push", len(paths))
		if err := deleteAll(v.store, paths); err != nil {
			return err
		}
	}

	return nil
}

// Pull events from remote.
// Does NOT require Unlock.
func (v *Vault) Pull(ctx context.Context) error {
	v.mtx.Lock()
	defer v.mtx.Unlock()
	return v.pull(ctx)
}

func (v *Vault) pull(ctx context.Context) error {
	if v.remote == nil {
		return errors.Errorf("no vault remote set")
	}
	if v.rk == nil {
		return errors.Errorf("no remote key")
	}

	index, err := v.index()
	if err != nil {
		return err
	}

	logger.Infof("Pulling vault items")
	vault, err := v.remote.Vault(ctx, v.rk, client.VaultIndex(index))
	if err != nil {
		return err
	}
	return v.saveRemoveVault(vault)
}

func (v *Vault) saveRemoveVault(vault *client.Vault) error {
	if vault == nil {
		return errors.Errorf("vault not found")
	}
	for _, event := range vault.Events {
		logger.Debugf("Pull %s", event.Path)
		if event.Path == "" {
			return errors.Errorf("invalid event (no path)")
		}
		if err := v.checkNonce(event.Nonce); err != nil {
			return err
		}

		if len(event.Data) == 0 {
			logger.Debugf("Deleting %s", event.Path)
			if _, err := v.store.Delete(event.Path); err != nil {
				return err
			}
		} else {
			logger.Debugf("Setting %s", event.Path)
			if err := v.store.Set(event.Path, event.Data); err != nil {
				return err
			}
		}

		ppath := ds.Path("pull", pad(event.Index), event.Path)
		eb, err := msgpack.Marshal(event)
		if err != nil {
			return err
		}
		if err := v.store.Set(ppath, eb); err != nil {
			return err
		}

		if err := v.commitNonce(event.Nonce); err != nil {
			return err
		}
	}

	// Update index
	if err := v.setIndex(vault.Index); err != nil {
		return err
	}

	return nil
}

// Spew to out.
func (v *Vault) Spew(prefix string, out io.Writer) error {
	docs, err := v.store.Documents(ds.Prefix(prefix))
	if err != nil {
		return err
	}
	if _, err := out.Write([]byte(fmt.Sprintf("%s\n", spew.Sdump(docs)))); err != nil {
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
	v.mtx.Lock()
	defer v.mtx.Unlock()

	if v.inc == 0 || (v.inc+n) >= v.incMax {
		inc, incMax, err := increment(v.store, ds.Path("db", "increment"), 1000)
		if err != nil {
			return "", err
		}
		v.inc, v.incMax = inc, incMax
	}
	v.inc = v.inc + n
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
		i, err := strconv.ParseInt(string(b), 10, 64)
		if err != nil {
			reterr = err
			return
		}
		inc = int64(i)
	}

	max := inc + size
	if err := st.Set(path, []byte(strconv.FormatInt(max, 10))); err != nil {
		reterr = err
		return
	}

	incMax = inc + size - 1
	return
}

func pad(n int64) string {
	if n > 999999999999999 {
		panic("int too large for padding")
	}
	return fmt.Sprintf("%015d", n)
}

// IsEmpty returns true if vault is empty.
func (v *Vault) IsEmpty() (bool, error) {
	docs, err := v.store.Documents(ds.Limit(1))
	if err != nil {
		return false, err
	}
	return len(docs) == 0, nil
}
