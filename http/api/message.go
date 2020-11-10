package api

import (
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore/events"
)

// MessagesResponse ...
type MessagesResponse struct {
	Messages []*events.Event `json:"msgs"`
	Index    int64           `json:"idx"`
}

// Message is encrypted by clients.
type Message struct {
	ID        string    `json:"id,omitempty" msgpack:"id,omitempty"`
	Prev      string    `json:"prev,omitempty" msgpack:"prev,omitempty"`
	Content   *Content  `json:"content,omitempty" msgpack:"content,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty" msgpack:"createdAt,omitempty"`
	// UpdatedAt time.Time `json:"updatedAt,omitempty" msgpack:"updatedAt,omitempty"`

	// Sender set from decrypt.
	Sender keys.ID `json:"-" msgpack:"-"`

	// RemoteIndex is set from the remote events API (untrusted).
	RemoteIndex int64 `json:"-" msgpack:"-"`
	// RemoteTimestamp is set from the remote events API (untrusted).
	RemoteTimestamp time.Time `json:"-" msgpack:"-"`
}

// ContentType is type for content.
type ContentType string

// Content types.
const (
	BinaryContent ContentType = "binary"
	UTF8Content   ContentType = "utf8"
)

// Content for message.
type Content struct {
	Data []byte      `json:"data,omitempty" msgpack:"data,omitempty"`
	Type ContentType `json:"type,omitempty" msgpack:"type,omitempty"`
}
