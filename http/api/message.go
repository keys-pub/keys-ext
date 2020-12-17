package api

import (
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/encoding"
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
	ID        string  `json:"id,omitempty" msgpack:"id,omitempty"`
	Prev      string  `json:"prev,omitempty" msgpack:"prev,omitempty"`
	Timestamp int64   `json:"ts,omitempty" msgpack:"ts,omitempty"`
	Sender    keys.ID `json:"sender" msgpack:"sender"`

	// For message text (optional).
	Text string `json:"text,omitempty" msgpack:"text,omitempty"`

	// ChannelInfo sets info.
	ChannelInfo *ChannelInfo `json:"channelInfo,omitempty" msgpack:"channelInfo,omitempty"`

	// Actions
	ChannelJoin  *ChannelJoin  `json:"channelJoin,omitempty" msgpack:"channelAccept,omitempty"`
	ChannelLeave *ChannelLeave `json:"channelLeave,omitempty" msgpack:"channelLeave,omitempty"`

	// ChannelInvite to invite to a new channel.
	ChannelInvite *ChannelInvite `json:"channelInvite,omitempty" msgpack:"channelInvite,omitempty"`

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

// Encrypt message.
// Experimental!
func (m *Message) Encrypt(sender *keys.EdX25519Key, recipient keys.ID) ([]byte, error) {
	if m.Sender == "" {
		return nil, errors.Errorf("message sender not set")
	}
	if m.Sender != sender.ID() {
		return nil, errors.Errorf("message sender mismatch")
	}
	b, err := msgpack.Marshal(m)
	if err != nil {
		return nil, err
	}
	signed := sender.Sign(b)

	pk := api.NewKey(recipient).AsX25519Public()
	if pk == nil {
		return nil, errors.Errorf("invalid message recipient")
	}
	encrypted := keys.CryptoBoxSeal(signed, pk)
	return encrypted, nil
}

// DecryptMessage decrypts message.
// Experimental!
func DecryptMessage(b []byte, key *keys.EdX25519Key) (*Message, error) {
	decrypted, err := keys.CryptoBoxSealOpen(b, key.X25519Key())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt message")
	}
	var message Message
	if err := msgpack.Unmarshal(decrypted[keys.SignOverhead:], &message); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal message")
	}

	pk := api.NewKey(message.Sender).AsEdX25519Public()
	if _, err := pk.Verify(decrypted); err != nil {
		return nil, err
	}

	return &message, nil
}

// DecryptMessageFromEvent decrypts a remote Event from Messages.
func DecryptMessageFromEvent(event *events.Event, key *keys.EdX25519Key) (*Message, error) {
	message, err := DecryptMessage(event.Data, key)
	if err != nil {
		return nil, err
	}
	message.RemoteIndex = event.Index
	message.RemoteTimestamp = event.Timestamp
	return message, nil
}

// NewMessageForChannelInfo ...
func NewMessageForChannelInfo(sender keys.ID, info *ChannelInfo) *Message {
	msg := NewMessage(sender)
	msg.ChannelInfo = info
	return msg
}

// NewMessageForChannelInvite ...
func NewMessageForChannelInvite(sender keys.ID, invite *ChannelInvite) *Message {
	msg := NewMessage(sender)
	msg.ChannelInvite = invite
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
