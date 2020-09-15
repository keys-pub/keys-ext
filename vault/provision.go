package vault

import (
	"sort"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Provision is unencrypted provision and parameters used by client auth.
type Provision struct {
	ID        string    `msgpack:"id" json:"id"`
	Type      AuthType  `msgpack:"type" json:"type"`
	CreatedAt time.Time `msgpack:"cts" json:"cts"`

	// AAGUID (for FIDO2HMACSecret)
	AAGUID string `msgpack:"aaguid,omitempty" json:"aaguid"`
	// Salt (for FIDO2HMACSecret)
	Salt []byte `msgpack:"salt,omitempty" json:"salt"`
	// NoPin (for FIDO2HMACSecret)
	NoPin bool `msgpack:"nopin,omitempty" json:"nopin"`
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
		if err := v.provisionSave(provision); err != nil {
			return err
		}
	}

	return nil
}

// Provisions are currently provisioned auth.
// Doesn't require Unlock().
func (v *Vault) Provisions() ([]*Provision, error) {
	path := docs.Path("provision")
	docs, err := v.store.Documents(docs.Prefix(path))
	if err != nil {
		return nil, err
	}
	provisions := []*Provision{}
	for _, doc := range docs {
		var provision Provision
		if err := msgpack.Unmarshal(doc.Data, &provision); err != nil {
			return nil, err
		}
		provisions = append(provisions, &provision)
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

	if _, err = v.provisionDelete(id); err != nil {
		return false, err
	}

	return v.authDelete(id)
}

// ProvisionSave for auth methods that need to store registration data before
// key is available (for example, FIDO2 hmac-secret).
func (v *Vault) ProvisionSave(provision *Provision) error {
	return v.provisionSave(provision)
}

// provision loads provision for id.
func (v *Vault) provision(id string) (*Provision, error) {
	logger.Debugf("Loading provision %s", id)
	path := docs.Path("provision", id)
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

// provisionSave saves provision.
func (v *Vault) provisionSave(provision *Provision) error {
	logger.Debugf("Saving provision %s", provision.ID)
	b, err := msgpack.Marshal(provision)
	if err != nil {
		return err
	}
	if err := v.set(docs.Path("provision", provision.ID), b, true); err != nil {
		return err
	}
	return nil
}

// provisionDelete removes provision.
func (v *Vault) provisionDelete(id string) (bool, error) {
	logger.Debugf("Deleting provision %s", id)
	path := docs.Path("provision", id)
	ok, err := v.store.Delete(path)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	if err := v.addToPush(docs.Path("provision", id), nil); err != nil {
		return true, err
	}
	return true, nil
}

func (v *Vault) isLastAuth(id string) (bool, error) {
	provisions, err := v.Provisions()
	if err != nil {
		return false, err
	}
	if len(provisions) == 1 && provisions[0].ID == id {
		return true, nil
	}
	return false, nil
}
