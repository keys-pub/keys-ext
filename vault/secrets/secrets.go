package secrets

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

// Secret to keep.
type Secret struct {
	ID   string `json:"id"`
	Type Type   `json:"type"`

	Name string `json:"name"`

	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	URL   string `json:"url,omitempty"`
	Notes string `json:"notes,omitempty"`

	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// Type types for secret.
type Type string

// Types for Secret.
const (
	UnknownType  Type = ""
	PasswordType Type = "password"
	NoteType     Type = "note"
)

func newSecretID() string {
	return encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
}

func newSecret() *Secret {
	return &Secret{
		ID:        newSecretID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewPassword creates a new password secret.
func NewPassword(name string, username string, password string, url string) *Secret {
	secret := newSecret()
	secret.Type = PasswordType
	secret.Name = name
	secret.Username = username
	secret.Password = password
	secret.URL = url
	return secret
}

// Save a secret.
// Returns true if secret was updated.
func Save(v *vault.Vault, secret *Secret) (*Secret, bool, error) {
	if secret == nil {
		return nil, false, errors.Errorf("nil secret")
	}

	if secret.ID == "" {
		secret.ID = newSecretID()
	}

	item, err := v.Get(secret.ID)
	if err != nil {
		return nil, false, err
	}

	updated := false
	if item != nil {
		secret.UpdatedAt = v.Now()
		item.Data = marshalSecret(secret)
		if err := v.Set(item); err != nil {
			return nil, false, err
		}
		updated = true
	} else {
		now := v.Now()
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

// Get a secret.
func Get(v *vault.Vault, id string) (*Secret, error) {
	item, err := v.Get(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return asSecret(item)
}

// SortDirection direction for sorting.
type SortDirection string

const (
	// Ascending direction.
	Ascending SortDirection = "asc"
	// Descending direction.
	Descending SortDirection = "desc"
)

// Options ...
type Options struct {
	Query         string
	Sort          string
	SortDirection SortDirection
}

// Option ...
type Option func(*Options)

func newSecretsOptions(opts ...Option) Options {
	var options Options
	for _, o := range opts {
		o(&options)
	}
	return options
}

// List ...
func List(v *vault.Vault, opt ...Option) ([]*Secret, error) {
	opts := newSecretsOptions(opt...)
	items, err := v.Items()
	if err != nil {
		return nil, err
	}
	query := strings.TrimSpace(opts.Query)
	out := make([]*Secret, 0, len(items))
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

// WithQuery ...
func WithQuery(q string) Option {
	return func(o *Options) { o.Query = q }
}

// WithSort ...
func WithSort(sort string) Option {
	return func(o *Options) { o.Sort = sort }
}

// WithSortDirection ...
func WithSortDirection(d SortDirection) Option {
	return func(o *Options) { o.SortDirection = d }
}

func secretsSort(secrets []*Secret, sortField string, sortDirection SortDirection, i, j int) bool {
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
func asSecret(item *vault.Item) (*Secret, error) {
	if item.Type != secretItemType {
		return nil, errors.Errorf("item type %s != %s", item.Type, secretItemType)
	}
	var secret Secret
	if err := json.Unmarshal(item.Data, &secret); err != nil {
		return nil, err
	}
	return &secret, nil
}

// secretItemType is type for secret.
const secretItemType string = "secret"

// newItem creates vault item for a secret.
func newItemForSecret(secret *Secret) (*vault.Item, error) {
	if secret.ID == "" {
		return nil, errors.Errorf("no secret id")
	}
	b := marshalSecret(secret)
	return vault.NewItem(secret.ID, b, secretItemType, secret.CreatedAt), nil
}

func marshalSecret(secret *Secret) []byte {
	b, err := json.Marshal(secret)
	if err != nil {
		panic(err)
	}
	return b
}
