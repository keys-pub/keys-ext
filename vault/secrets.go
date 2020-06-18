package vault

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/keys-pub/keys/secret"
	"github.com/pkg/errors"
)

// SaveSecret saves a secret.
// Returns true if secret was updated.
func (v *Vault) SaveSecret(secret *secret.Secret) (*secret.Secret, bool, error) {
	if secret == nil {
		return nil, false, errors.Errorf("no secret")
	}

	if strings.TrimSpace(secret.ID) == "" {
		return nil, false, errors.Errorf("no secret id")
	}

	item, err := v.Get(secret.ID)
	if err != nil {
		return nil, false, err
	}

	updated := false
	if item != nil {
		secret.UpdatedAt = v.clock()
		item.Data = marshalSecret(secret)
		if err := v.Set(item); err != nil {
			return nil, false, err
		}
		updated = true
	} else {
		now := v.clock()
		secret.CreatedAt = now
		secret.UpdatedAt = now

		item, err := newItemForSecret(secret)
		if err != nil {
			return nil, false, err
		}
		if err := v.Set(item); err != nil {
			return nil, false, err
		}
	}

	return secret, updated, nil
}

// Secret for ID.
func (v *Vault) Secret(id string) (*secret.Secret, error) {
	item, err := v.Get(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return asSecret(item)
}

// Secrets ...
func (v *Vault) Secrets() ([]*secret.Secret, error) {
	items, err := v.Items()
	if err != nil {
		return nil, err
	}
	out := make([]*secret.Secret, 0, len(items))
	for _, item := range items {
		if item.Type != secretItemType {
			continue
		}
		key, err := asSecret(item)
		if err != nil {
			return nil, err
		}
		out = append(out, key)
	}
	logger.Debugf("Found %d secrets", len(out))
	return out, nil
}

// asSecret returns Secret for Item.
func asSecret(item *Item) (*secret.Secret, error) {
	if item.Type != secretItemType {
		return nil, errors.Errorf("item type %s != %s", item.Type, secretItemType)
	}
	var secret secret.Secret
	if err := json.Unmarshal(item.Data, &secret); err != nil {
		return nil, err
	}
	return &secret, nil
}

// secretItemType is type for secret.
const secretItemType string = "secret"

// newItem creates vault item for a secret.
func newItemForSecret(secret *secret.Secret) (*Item, error) {
	if secret.ID == "" {
		return nil, errors.Errorf("no secret id")
	}
	b := marshalSecret(secret)
	return NewItem(secret.ID, b, secretItemType, time.Now()), nil
}

func marshalSecret(secret *secret.Secret) []byte {
	b, err := json.Marshal(secret)
	if err != nil {
		panic(err)
	}
	return b
}
