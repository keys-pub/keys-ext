package vault

import (
	"strings"

	"github.com/keys-pub/keys/ds"
	"github.com/vmihailenco/msgpack/v4"
)

// ItemHistory returns history of an item.
// Items with empty data are deleted items.
// This is slow.
func (v *Vault) ItemHistory(id string) ([]*Item, error) {
	path := ds.Path("pull")
	docs, err := v.store.Documents(ds.Prefix(path), ds.NoData())
	if err != nil {
		return nil, err
	}
	paths := []string{}
	for _, doc := range docs {
		if strings.HasPrefix(ds.PathFrom(doc.Path, 2), ds.Path("item", id)) {
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
		var event ds.Event
		if err := msgpack.Unmarshal(b, &event); err != nil {
			return nil, err
		}
		item, err := decryptItem(event.Data, v.mk)
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
	path := ds.Path("push")
	docs, err := v.store.Documents(ds.Prefix(path))
	if err != nil {
		return nil, err
	}
	items := []*Item{}
	for _, doc := range docs {
		pc := ds.PathComponents(doc.Path)
		if pc[2] != "item" || pc[3] != id {
			continue
		}
		item, err := decryptItem(doc.Data, v.mk)
		if err != nil {
			return nil, err
		}
		if item.ID == id {
			items = append(items, item)
		}
	}
	return items, nil
}
