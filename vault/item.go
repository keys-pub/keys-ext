package vault

import (
	"time"

	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Item in the vault.
type Item struct {
	ID   string `msgpack:"id"`
	Data []byte `msgpack:"dat"`

	// Type for item data.
	Type string `msgpack:"typ,omitempty"`

	// CreatedAt when item was created.
	CreatedAt time.Time `msgpack:"cts,omitempty"`

	// Timestamp is set from the remote.
	// This can be used for versioning.
	Timestamp time.Time `msgpack:"ts,omitempty"`
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
	if err := checkItemSize(item); err != nil {
		return nil, err
	}
	b, err := msgpack.Marshal(item)
	if err != nil {
		return nil, err
	}
	out := secretBoxSeal(b, mk)

	if len(out) > maxSize {
		return nil, ErrItemValueTooLarge
	}

	return out, nil
}

func decryptItem(b []byte, mk *[32]byte) (*Item, error) {
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
	return &item, nil
}

const maxID = 254
const maxType = 32
const maxData = 2048

// maxSize (windows credential blob)
const maxSize = (5 * 512)

// ErrItemValueTooLarge is item value is too large.
// Item.ID is max of 254 bytes.
// Item.Type is max of 32 bytes.
// Item.Data is max of 2048 bytes.
var ErrItemValueTooLarge = errors.New("item value is too large")

func checkItemSize(item *Item) error {
	if len(item.ID) > maxID {
		return ErrItemValueTooLarge
	}
	if len(item.Type) > maxType {
		return ErrItemValueTooLarge
	}
	if len(item.Data) > maxData {
		return ErrItemValueTooLarge
	}
	return nil
}
