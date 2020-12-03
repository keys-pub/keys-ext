package vault

import (
	"time"

	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Item in the vault, uses msgpack.
type Item struct {
	ID   string `msgpack:"id"`
	Data []byte `msgpack:"dat"`

	// Type for item data.
	Type string `msgpack:"typ,omitempty"`

	// CreatedAt when item was created.
	// TODO: Updated at
	CreatedAt time.Time `msgpack:"cts,omitempty"`
}

// NewItem creates an item.
func NewItem(id string, b []byte, typ string, createdAt time.Time) *Item {
	return &Item{
		ID:        id,
		Data:      b,
		Type:      typ,
		CreatedAt: createdAt,
	}
}

// Encrypt item.
func (i *Item) Encrypt(mk *[32]byte) ([]byte, error) {
	return encryptItem(i, mk)
}

func encryptItem(item *Item, mk *[32]byte) ([]byte, error) {
	if mk == nil {
		return nil, ErrLocked
	}
	if item.ID == "" {
		return nil, errors.Errorf("invalid id")
	}
	if len(item.Data) > 32*1024 {
		return nil, ErrItemValueTooLarge
	}
	b, err := msgpack.Marshal(item)
	if err != nil {
		return nil, err
	}
	out := secretBoxSeal(b, mk)

	return out, nil
}

func decryptItem(b []byte, mk *[32]byte, ad string) (*Item, error) {
	if mk == nil {
		return nil, ErrLocked
	}
	decrypted, ok := secretBoxOpen(b, mk)
	if !ok {
		return nil, ErrInvalidAuth
	}
	var item Item
	if err := msgpack.Unmarshal(decrypted, &item); err != nil {
		return nil, err
	}
	if ad != "" && item.ID != ad {
		return nil, errors.Errorf("invalid associated data")
	}
	return &item, nil
}

// ErrItemValueTooLarge is item value is too large.
var ErrItemValueTooLarge = errors.New("item value is too large")
