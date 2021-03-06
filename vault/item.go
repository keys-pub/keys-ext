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

	// Timestamp for item.
	Timestamp time.Time `msgpack:"cts,omitempty"`

	// TODO: Specify prev item (for chaining).
}

// NewItem creates an item.
// Item IDs are NOT encrypted locally and are provided for fast lookups.
func NewItem(id string, b []byte, typ string, ts time.Time) *Item {
	return &Item{
		ID:        id,
		Data:      b,
		Type:      typ,
		Timestamp: ts,
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
	if b == nil {
		return nil, errors.Errorf("nothing to decrypt")
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
