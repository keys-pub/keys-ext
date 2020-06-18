package vault

import (
	"sort"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Provision is unencrypted provision and parameters used by client auth.
type Provision struct {
	ID        string    `msgpack:"id"`
	Type      AuthType  `msgpack:"type"`
	CreatedAt time.Time `msgpack:"cts"`

	// AAGUID (for FIDO2HMACSecret)
	AAGUID string `msgpack:"aaguid"`
	// Salt (for FIDO2HMACSecret)
	Salt []byte `msgpack:"salt"`
	// NoPin (for FIDO2HMACSecret)
	NoPin bool `msgpack:"nopin"`
}

// NewProvision creates a new provision.
func NewProvision(typ AuthType) *Provision {
	return &Provision{
		ID:        encoding.MustEncode(keys.RandBytes(32), encoding.Base62),
		Type:      typ,
		CreatedAt: time.Now(),
	}
}

// Provision new auth.
// Requires Unlock().
func (v *Vault) Provision(key *[32]byte, provision *Provision) error {
	if v.mk == nil {
		return ErrLocked
	}
	if provision == nil {
		return errors.Errorf("no provision")
	}

	if err := v.authCreate(provision.ID, key, v.mk); err != nil {
		return err
	}

	if provision != nil {
		if err := v.saveProvision(provision); err != nil {
			return err
		}
	}

	return nil
}

// Provisions are currently provisioned auth.
// Doesn't require Unlock().
func (v *Vault) Provisions() ([]*Provision, error) {
	path := v.protocol.Path(provisionEntity)
	iter, err := v.store.Documents(ds.Prefix(path))
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	provisions := []*Provision{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if doc == nil {
			break
		}
		var provision Provision
		if err := msgpack.Unmarshal(doc.Data, &provision); err != nil {
			return nil, err
		}
		provisions = append(provisions, &provision)
	}

	// Check for v0 auth
	v0, err := v.store.Exists("#auth")
	if err != nil {
		return nil, err
	}
	if v0 {
		provisions = append(provisions, &Provision{
			ID:        "auth.v0",
			CreatedAt: time.Time{},
		})
	}

	// Sort by time
	sort.Slice(provisions, func(i, j int) bool { return provisions[i].CreatedAt.Before(provisions[j].CreatedAt) })

	return provisions, nil
}

// Deprovision auth.
// Doesn't require Unlock().
func (v *Vault) Deprovision(id string, force bool) (bool, error) {
	logger.Debugf("Deprovisioning %s", id)

	// Check to make sure not the last provision (unless forced)
	last, err := v.isLastAuth(id)
	if err != nil {
		return false, err
	}
	if !force && last {
		return false, errors.Errorf("failed to deprovision: last auth")
	}

	// Check for v0 auth
	if id == "auth.v0" {
		return v.store.Delete("#auth")
	}

	if _, err = v.deleteProvision(id); err != nil {
		return false, err
	}

	return v.authDelete(id)
}

// SaveProvision for auth methods that need to store registration data before
// key is available (for example, FIDO2 hmac-secret).
func (v *Vault) SaveProvision(provision *Provision) error {
	return v.saveProvision(provision)
}

// loadProvision loads provision for id.
func (v *Vault) loadProvision(id string) (*Provision, error) {
	logger.Debugf("Loading provision %s", id)
	path := v.protocol.Path(provisionEntity, id)
	b, err := v.store.Get(path)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	var provision Provision
	if err := msgpack.Unmarshal(b, &provision); err != nil {
		return nil, err
	}
	return &provision, nil
}

// saveProvision saves provision.
func (v *Vault) saveProvision(provision *Provision) error {
	logger.Debugf("Saving provision %s", provision.ID)
	path := v.protocol.Path(provisionEntity, provision.ID)
	b, err := msgpack.Marshal(provision)
	if err != nil {
		return err
	}
	if err := v.store.Set(path, b); err != nil {
		return err
	}
	return nil
}

// deleteProvision removes provision.
func (v *Vault) deleteProvision(id string) (bool, error) {
	logger.Debugf("Deleting provision %s", id)

	path := v.protocol.Path(provisionEntity, id)
	return v.store.Delete(path)
}

func (v *Vault) isLastAuth(id string) (bool, error) {
	provisions, err := v.Provisions()
	if err != nil {
		return false, err
	}
	if len(provisions) == 0 && id == "auth.v0" {
		return true, nil
	}
	if len(provisions) == 1 && provisions[0].ID == id {
		return true, nil
	}
	return false, nil
}
