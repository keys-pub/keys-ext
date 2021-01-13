package vault

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	httpclient "github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// ErrNotOpen if you try to use the vault when it isn't open.
var ErrNotOpen = errors.Errorf("vault not open")

// ErrAlreadyOpen if you try to open when it is already open.
var ErrAlreadyOpen = errors.Errorf("vault already open")

// Vault stores keys and secrets.
type Vault struct {
	mtx sync.Mutex

	store  Store
	client *httpclient.Client
	clock  tsutil.Clock

	mk     *[32]byte
	remote *Remote

	auto *time.Timer

	checkedAt time.Time
	checkMtx  sync.Mutex
}

// New vault.
func New(st Store, opt ...Option) *Vault {
	opts := newOptions(opt...)
	return &Vault{
		store: st,
		clock: opts.Clock,
	}
}

// Now is current time.
func (v *Vault) Now() time.Time {
	return v.clock.Now()
}

// Store used by vault.
func (v *Vault) Store() Store {
	return v.store
}

// Open vault.
func (v *Vault) Open() error {
	logger.Infof("Open %s", v.store.Path())
	if err := v.store.Open(); err != nil {
		return errors.Wrapf(err, "failed to open vault")
	}
	return nil
}

// Close vault.
func (v *Vault) Close() error {
	if v.auto != nil {
		v.auto.Stop()
		v.auto = nil
	}
	logger.Infof("Close %s", v.store.Path())
	// TODO: Sync could still be running when we close, this might be
	//       ok, since it will error and eventually stop?
	if err := v.store.Close(); err != nil {
		return errors.Wrapf(err, "failed to close vault")
	}
	return nil
}

// Reset vault.
func (v *Vault) Reset() error {
	status, err := v.Status()
	if err != nil {
		return err
	}
	if status == Unlocked {
		return errors.Errorf("failed to reset: vault is unlocked")
	}
	if err := v.store.Reset(); err != nil {
		return errors.Wrapf(err, "failed to reset vault")
	}
	return nil
}

// SetClient sets the client.
func (v *Vault) SetClient(client *httpclient.Client) {
	v.client = client
}

// setMasterKey sets the master key.
func (v *Vault) setMasterKey(mk *[32]byte) error {
	if err := v.setAuthFromMasterKey(mk); err != nil {
		return err
	}
	v.mk = mk
	return nil
}

func (v *Vault) setAuthFromMasterKey(mk *[32]byte) error {
	rsalt, err := v.getRemoteSalt(true)
	if err != nil {
		return err
	}

	// Derive API key
	seed := keys.Bytes32(keys.HKDFSHA256(mk[:], 32, rsalt, []byte("keys.pub/rk")))
	rk := keys.NewEdX25519KeyFromSeed(seed)

	remote := &Remote{Key: rk, Salt: rsalt}

	// If auth was already set in Clone, we should double check it matches the
	// auth generated from the master key.
	if v.remote != nil {
		if !reflect.DeepEqual(v.remote, remote) {
			return errors.Errorf("remote auth is different than expected")
		}
	}
	v.remote = remote

	return nil
}

func (v *Vault) clearRemote() error {
	v.remote = nil
	if err := v.setRemoteSalt(nil); err != nil {
		return err
	}
	if v.mk != nil {
		if err := v.setAuthFromMasterKey(v.mk); err != nil {
			return err
		}
	}
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

// Set vault item.
func (v *Vault) Set(item *Item) error {
	return v.setItem(item, true)
}

func (v *Vault) setItem(item *Item, addToPush bool) error {
	b, err := encryptItem(item, v.mk)
	if err != nil {
		return err
	}
	path := dstore.Path("item", item.ID)
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
	inc, err := v.pushIndexNext()
	if err != nil {
		return err
	}
	push := dstore.Path("push", pad(inc), path)
	if err := v.store.Set(push, b); err != nil {
		return err
	}

	if v.auto != nil {
		v.auto.Stop()
		v.auto = nil
	}
	v.auto = time.AfterFunc(time.Second*2, func() {
		if _, err := v.CheckSync(context.TODO(), time.Duration(0)); err != nil {
			logger.Errorf("Failed to check sync: %v", err)
		}
	})

	return nil
}

// Get vault item.
func (v *Vault) Get(id string) (*Item, error) {
	if id == "" {
		return nil, errors.Errorf("empty id")
	}
	path := dstore.Path("item", id)
	b, err := v.store.Get(path)
	if err != nil {
		return nil, err
	}
	if b == nil {
		logger.Debugf("Path not found %s", path)
		return nil, nil
	}
	item, err := decryptItem(b, v.mk, id)
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

// Items to list.
func (v *Vault) Items() ([]*Item, error) {
	path := dstore.Path("item")
	ds, err := v.store.List(&ListOptions{Prefix: path})
	if err != nil {
		return nil, err
	}
	items := []*Item{}
	for _, doc := range ds {
		id := dstore.PathLast(doc.Path)
		item, err := decryptItem(doc.Data, v.mk, id)
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
	if v.client == nil {
		return errors.Errorf("no vault client set")
	}
	if v.remote == nil {
		return errors.Errorf("no remote set")
	}

	paths := []string{}
	events := []*httpclient.VaultEvent{}

	// Get events from push.
	path := dstore.Path("push")
	ds, err := v.store.List(&ListOptions{Prefix: path})
	if err != nil {
		return err
	}

	for _, doc := range ds {
		logger.Debugf("Push %s", doc.Path)
		paths = append(paths, doc.Path)
		path := dstore.PathFrom(doc.Path, 2)
		event := &httpclient.VaultEvent{Path: path, Data: doc.Data}
		events = append(events, event)
	}

	if len(events) > 0 {
		logger.Infof("Pushing %d vault events", len(events))
		if err := v.client.VaultSend(ctx, v.remote.Key, events); err != nil {
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
	// Keep pulling until no more.
	for {
		truncated, err := v.pullNext(ctx)
		if err != nil {
			return err
		}
		if !truncated {
			break
		}
	}
	return nil
}

func (v *Vault) pullNext(ctx context.Context) (bool, error) {
	if v.client == nil {
		return false, errors.Errorf("no vault client set")
	}
	if v.remote == nil {
		return false, errors.Errorf("no remote set")
	}

	index, err := v.pullIndex()
	if err != nil {
		return false, err
	}

	logger.Infof("Pulling vault items")
	vault, err := v.client.Vault(ctx, v.remote.Key, httpclient.VaultIndex(index))
	if err != nil {
		return false, err
	}
	if err := v.saveRemoteVault(vault); err != nil {
		return false, err
	}

	return vault.Truncated, nil
}

func (v *Vault) saveRemoteVault(vault *httpclient.Vault) error {
	if vault == nil {
		return errors.Errorf("vault not found")
	}

	for _, event := range vault.Events {
		logger.Debugf("Pull %s", event.Path)
		if event.Path == "" {
			return errors.Errorf("invalid event (no path)")
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

		pull := dstore.Path("pull", pad(event.RemoteIndex), event.Path)
		eb, err := msgpack.Marshal(event)
		if err != nil {
			return err
		}
		if err := v.store.Set(pull, eb); err != nil {
			return err
		}
	}

	// Update pull index.
	if err := v.setPullIndex(vault.Index); err != nil {
		return err
	}

	return nil
}

// Spew to out.
func (v *Vault) Spew(prefix string, out io.Writer) error {
	docs, err := v.store.List(&ListOptions{Prefix: prefix})
	if err != nil {
		return err
	}
	if _, err := out.Write([]byte(fmt.Sprintf("%s\n", spew.Sdump(docs)))); err != nil {
		return err
	}
	return nil
}

// IsEmpty returns true if vault is empty.
func (v *Vault) IsEmpty() (bool, error) {
	docs, err := v.store.List(&ListOptions{Limit: 1})
	if err != nil {
		return false, err
	}
	return len(docs) == 0, nil
}
