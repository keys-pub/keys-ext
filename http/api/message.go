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

	// Notifications
	ChannelInvites *ChannelInvitesNn `json:"channelInvites,omitempty" msgpack:"channelInvites,omitempty"`
	ChannelJoin    *ChannelJoinNn    `json:"channelJoin,omitempty" msgpack:"channelAccept,omitempty"`
	ChannelLeave   *ChannelLeaveNn   `json:"channelLeave,omitempty" msgpack:"channelLeave,omitempty"`

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
func NewMessage(sender keys.ID) *Message {
	return &Message{
		ID:        NewID(),
		Sender:    sender,
		Timestamp: tsutil.Millis(time.Now()),
	}
}

// WithPrev ...
func (m *Message) WithPrev(prev string) *Message {
	m.Prev = prev
	return m
}

// WithText ...
func (m *Message) WithText(text string) *Message {
	m.Text = text
	return m
}

// WithTimestamp ...
func (m *Message) WithTimestamp(ts int64) *Message {
	m.Timestamp = ts
	return m
}

// NewMessageForChannelInfo ...
func NewMessageForChannelInfo(sender keys.ID, info *ChannelInfo) *Message {
	msg := NewMessage(sender)
	msg.ChannelInfo = info
	return msg
}

// NewMessageForChannelInvites ...
func NewMessageForChannelInvites(sender keys.ID, users ...keys.ID) *Message {
	msg := NewMessage(sender)
	msg.ChannelInvites = &ChannelInvitesNn{Users: users}
	return msg
}

// NewMessageForChannelJoin ...
func NewMessageForChannelJoin(sender keys.ID, user keys.ID) *Message {
	msg := NewMessage(sender)
	msg.ChannelJoin = &ChannelJoinNn{User: user}
	return msg
}

// NewMessageForChannelLeave ...
func NewMessageForChannelLeave(sender keys.ID, user keys.ID) *Message {
	msg := NewMessage(sender)
	msg.ChannelLeave = &ChannelLeaveNn{User: user}
	return msg
}
