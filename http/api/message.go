package api

import (
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
)

// MessagesResponse ...
type MessagesResponse struct {
	Messages  []*events.Event `json:"msgs"`
	Index     int64           `json:"idx"`
	Truncated bool            `json:"truncated,omitempty"`
}

// Message is encrypted by clients.
type Message struct {
	ID        string `json:"id,omitempty" msgpack:"id,omitempty"`
	Prev      string `json:"prev,omitempty" msgpack:"prev,omitempty"`
	Timestamp int64  `json:"ts,omitempty" msgpack:"ts,omitempty"`
	// UpdatedAt time.Time `json:"updatedAt,omitempty" msgpack:"updatedAt,omitempty"`

	// For message text (optional).
	Text string `json:"text,omitempty" msgpack:"text,omitempty"`

	// For channel info (optional).
	ChannelInfo *ChannelInfo `json:"channelInfo,omitempty" msgpack:"channelInfo,omitempty"`

	// Sender set from decrypt.
	Sender keys.ID `json:"-" msgpack:"-"`

	// RemoteIndex is set from the remote events API (untrusted).
	RemoteIndex int64 `json:"-" msgpack:"-"`
	// RemoteTimestamp is set from the remote events API (untrusted).
	RemoteTimestamp int64 `json:"-" msgpack:"-"`
}

// NewID returns a new random ID (string).
func NewID() string {
	return encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
}

// NewMessage creates a new empty message.
func NewMessage() *Message {
	return &Message{
		ID:        NewID(),
		Timestamp: tsutil.Millis(time.Now()),
	}
}
