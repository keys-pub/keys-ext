package vault

import (
	"strings"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/vmihailenco/msgpack/v4"
)

// ItemHistory returns history of an item.
// Items with empty data are deleted items.
// This is slow.
func (v *Vault) ItemHistory(id string) ([]*Item, error) {
	path := dstore.Path("pull")
	ds, err := v.store.List(&ListOptions{Prefix: path, NoData: true})
	if err != nil {
		return nil, err
	}
	paths := []string{}
	for _, doc := range ds {
		if strings.HasPrefix(dstore.PathFrom(doc.Path, 2), dstore.Path("item", id)) {
			paths = append(paths, doc.Path)
		}
	}

	items := make([]*Item, 0, len(paths))
	for _, p := range paths {
		b, err := v.store.Get(p)
		if err != nil {
			return nil, err
		}
		if b == nil {
			continue
		}
		var event api.Event
		if err := msgpack.Unmarshal(b, &event); err != nil {
			return nil, err
		}
		id := dstore.PathLast(p)
		item, err := decryptItem(event.Data, v.mk, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	pending, err := v.findPendingItems(id)
	if err != nil {
		return nil, err
	}
	items = append(items, pending...)

	return items, nil
}

// findPendingItems returns list of pending items awaiting push.
// Requires Unlock.
func (v *Vault) findPendingItems(id string) ([]*Item, error) {
	path := dstore.Path("push")
	entries, err := v.store.List(&ListOptions{Prefix: path})
	if err != nil {
		return nil, err
	}
	items := []*Item{}
	for _, entry := range entries {
		pc := dstore.PathComponents(entry.Path)
		if pc[2] != "item" || pc[3] != id {
			continue
		}
		item, err := decryptItem(entry.Data, v.mk, id)
		if err != nil {
			return nil, err
		}
		if item.ID == id {
			items = append(items, item)
		}
	}
	return items, nil
}
