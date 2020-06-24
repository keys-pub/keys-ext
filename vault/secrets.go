package vault

import (
	"encoding/json"
	"sort"
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
func (v *Vault) Secrets(opt ...SecretsOption) ([]*secret.Secret, error) {
	opts := newSecretsOptions(opt...)
	items, err := v.Items()
	if err != nil {
		return nil, err
	}
	query := strings.TrimSpace(opts.Query)
	out := make([]*secret.Secret, 0, len(items))
	for _, item := range items {
		if item.Type != secretItemType {
			continue
		}
		secret, err := asSecret(item)
		if err != nil {
			return nil, err
		}

		if query == "" ||
			strings.Contains(secret.Name, query) ||
			strings.Contains(secret.Username, query) ||
			strings.Contains(secret.URL, query) ||
			strings.Contains(secret.Notes, query) {
			out = append(out, secret)
		}
	}
	logger.Debugf("Found %d secrets", len(out))

	sortField := opts.Sort
	if sortField == "" {
		sortField = "name"
	}
	switch sortField {
	case "id", "name", "username":
	default:
		return nil, errors.Errorf("invalid sort field %s", sortField)
	}
	sortDirection := opts.SortDirection
	if sortDirection == "" {
		sortDirection = Ascending
	}
	sort.Slice(out, func(i, j int) bool {
		return secretsSort(out, sortField, sortDirection, i, j)
	})

	return out, nil
}

// Secrets options.
var Secrets = secrets{}

type secrets struct{}

func (s secrets) Query(q string) SecretsOption {
	return func(o *SecretsOptions) { o.Query = q }
}
func (s secrets) Sort(sort string) SecretsOption {
	return func(o *SecretsOptions) { o.Sort = sort }
}
func (s secrets) SortDirection(d SortDirection) SecretsOption {
	return func(o *SecretsOptions) { o.SortDirection = d }
}

func secretsSort(secrets []*secret.Secret, sortField string, sortDirection SortDirection, i, j int) bool {
	switch sortField {
	case "id":
		if sortDirection == Descending {
			return secrets[i].ID > secrets[j].ID
		}
		return secrets[i].ID < secrets[j].ID
	case "name":
		if secrets[i].Name == secrets[j].Name {
			return secretsSort(secrets, "id", sortDirection, i, j)
		}
		if sortDirection == Descending {
			return secrets[i].Name > secrets[j].Name
		}
		return secrets[i].Name < secrets[j].Name
	case "username":
		if secrets[i].Username == secrets[j].Username {
			return secretsSort(secrets, "id", sortDirection, i, j)
		}
		if sortDirection == Descending {
			return secrets[i].Username > secrets[j].Username
		}
		return secrets[i].Username < secrets[j].Username
	default:
		return secretsSort(secrets, "name", sortDirection, i, j)
	}
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
