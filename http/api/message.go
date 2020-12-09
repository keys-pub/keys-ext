package api

import (
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
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

	// Actions (optional).
	ChannelInvites   *ChannelInvites   `json:"channelInvites,omitempty" msgpack:"channelInvites,omitempty"`
	ChannelUninvites *ChannelUninvites `json:"channelUninvites,omitempty" msgpack:"channelUninvites,omitempty"`
	ChannelJoin      *ChannelJoin      `json:"channelJoin,omitempty" msgpack:"channelAccept,omitempty"`
	ChannelLeave     *ChannelLeave     `json:"channelLeave,omitempty" msgpack:"channelLeave,omitempty"`

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
	msg.ChannelInvites = &ChannelInvites{Users: users}
	return msg
}

// NewMessageForChannelUninvites ...
func NewMessageForChannelUninvites(sender keys.ID, users ...keys.ID) *Message {
	msg := NewMessage(sender)
	msg.ChannelUninvites = &ChannelUninvites{Users: users}
	return msg
}

// NewMessageForChannelJoin ...
func NewMessageForChannelJoin(sender keys.ID, user keys.ID) *Message {
	msg := NewMessage(sender)
	msg.ChannelJoin = &ChannelJoin{User: user}
	return msg
}

// NewMessageForChannelLeave ...
func NewMessageForChannelLeave(sender keys.ID, user keys.ID) *Message {
	msg := NewMessage(sender)
	msg.ChannelLeave = &ChannelLeave{User: user}
	return msg
}

// EncryptMessage encrypts a message.
func EncryptMessage(message *Message, sender *keys.EdX25519Key, channel keys.ID) ([]byte, error) {
	if message.Sender == "" {
		return nil, errors.Errorf("message sender not set")
	}
	if message.Sender != sender.ID() {
		return nil, errors.Errorf("message sender mismatch")
	}
	b, err := msgpack.Marshal(message)
	if err != nil {
		return nil, err
	}
	encrypted, err := saltpack.Signcrypt(b, false, sender, channel.ID())
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}

// DecryptMessage decrypts a remote Event from Messages.
func DecryptMessage(event *events.Event, kr saltpack.Keyring) (*Message, error) {
	decrypted, pk, err := saltpack.SigncryptOpen(event.Data, false, kr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt message")
	}
	var message Message
	if err := msgpack.Unmarshal(decrypted, &message); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal message")
	}
	message.Sender = pk.ID()
	message.RemoteIndex = event.Index
	message.RemoteTimestamp = event.Timestamp
	return &message, nil
}
