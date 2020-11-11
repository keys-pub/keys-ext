package vault

import (
	"strings"

	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// ConvertKeyring converts keyring store.
func ConvertKeyring(kr keyring.Keyring, to *Vault) (bool, error) {
	items, err := kr.Items("")
	if err != nil {
		return false, err
	}
	if len(items) == 0 {
		return false, nil
	}
	for _, item := range items {
		// #salt
		if item.ID == "#salt" {
			if err := to.set(dstore.Path("config", "salt"), item.Data, true); err != nil {
				return false, err
			}
			continue
		}

		// #auth
		if item.ID == "#auth" {
			if err := to.set(dstore.Path("auth", "v0"), item.Data, true); err != nil {
				return false, err
			}
			provision := &Provision{
				ID:   "v0",
				Type: PasswordAuth,
			}
			b, err := msgpack.Marshal(provision)
			if err != nil {
				return false, err
			}
			if err := to.set(dstore.Path("provision", "v0"), b, true); err != nil {
				return false, err
			}
			continue
		}

		spl := strings.Split(item.ID, "-")

		switch spl[0] {
		// #auth-
		case "#auth":
			if len(spl) < 2 {
				return false, errors.Errorf("unsupported id %s", item.ID)
			}
			if err := to.set(dstore.Path("auth", spl[1]), item.Data, true); err != nil {
				return false, err
			}
		// #provision-
		case "#provision":
			if len(spl) < 2 {
				return false, errors.Errorf("unsupported id %s", item.ID)
			}
			if err := to.set(dstore.Path("provision", spl[1]), item.Data, true); err != nil {
				return false, err
			}
		// items
		default:
			if strings.HasPrefix(item.ID, "#") {
				continue
			}
			if strings.HasPrefix(item.ID, ".") {
				continue
			}
			if err := to.set(dstore.Path("item", item.ID), item.Data, true); err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// convertID converts old IDs at runtime that we can't convert normally, such as
// auth item IDs.
func convertID(id string) string {
	if id == "#auth" {
		return "v0"
	}
	if strings.HasPrefix(id, "#auth-") {
		return id[6:]
	}
	return id
}
