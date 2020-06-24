package vault

import (
	"strings"

	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// ConvertKeyring converts keyring store.
func ConvertKeyring(kr keyring.Keyring, to *Vault) error {
	iter, err := kr.Documents()
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

		// #salt
		if doc.Path == "#salt" {
			if err := to.set(ds.Path("config", "salt"), doc.Data); err != nil {
				return err
			}
			continue
		}

		// #auth
		if doc.Path == "#auth" {
			if err := to.set(ds.Path("auth", "v0"), doc.Data); err != nil {
				return err
			}
			provision := &Provision{
				ID:        "v0",
				Type:      PasswordAuth,
				CreatedAt: doc.CreatedAt,
			}
			b, err := msgpack.Marshal(provision)
			if err != nil {
				return err
			}
			if err := to.set(ds.Path("provision", "v0"), b); err != nil {
				return err
			}
			continue
		}

		spl := strings.Split(doc.Path, "-")

		switch spl[0] {
		// #auth-
		case "#auth":
			if len(spl) < 2 {
				return errors.Errorf("unsupported id %s", doc.Path)
			}
			if err := to.set(ds.Path("auth", spl[1]), doc.Data); err != nil {
				return err
			}
		// #provision-
		case "#provision":
			if len(spl) < 2 {
				return errors.Errorf("unsupported id %s", doc.Path)
			}
			if err := to.set(ds.Path("provision", spl[1]), doc.Data); err != nil {
				return err
			}
		// items
		default:
			if strings.HasPrefix(doc.Path, "#") {
				continue
			}
			if err := to.set(ds.Path("item", doc.Path), doc.Data); err != nil {
				return err
			}
		}
	}

	return nil
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
