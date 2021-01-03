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

// Events ...
type Events struct {
	Events    []*events.Event `json:"events"`
	Index     int64           `json:"idx"`
	Truncated bool            `json:"truncated,omitempty"`
}

// Decrypt events into messages.
func (e Events) Decrypt(key *keys.EdX25519Key) ([]*Message, error) {
	msgs := make([]*Message, 0, len(e.Events))
	for _, event := range e.Events {
		msg, err := DecryptMessageFromEvent(event, key)
		if err != nil {
			// TODO: Skip invalid messages
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
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

	// ChannelInvites to invite to a new channel.
	ChannelInvites []*ChannelInvite `json:"channelInvites,omitempty" msgpack:"channelInvites,omitempty"`

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
	if m.RemoteTimestamp != 0 {
		return nil, errors.Errorf("remote timestamp should be omitted on send")
	}
	if m.RemoteIndex != 0 {
		return nil, errors.Errorf("remote index should be omitted on send")
	}
	if m.Timestamp == 0 {
		return nil, errors.Errorf("message timestamp is not set")
	}
	if m.Sender == "" {
		return nil, errors.Errorf("message sender not set")
	}
	if m.Sender != sender.ID() {
		return nil, errors.Errorf("message sender mismatch")
	}
	return Encrypt(m, sender, recipient)
}

// Encrypt marshals to msgpack, crypto_sign and then crypto_box_seal.
// We are doing sign then encrypt to hide the sender.
func Encrypt(i interface{}, sender *keys.EdX25519Key, recipient keys.ID) ([]byte, error) {
	b, err := msgpack.Marshal(i)
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
	var message Message
	sig, err := Decrypt(b, &message, key)
	if err != nil {
		return nil, err
	}
	pk := api.NewKey(message.Sender).AsEdX25519Public()
	if _, err := pk.Verify(sig); err != nil {
		return nil, err
	}
	return &message, nil
}

// Decrypt value, returning signature.
func Decrypt(b []byte, v interface{}, key *keys.EdX25519Key) ([]byte, error) {
	sig, err := keys.CryptoBoxSealOpen(b, key.X25519Key())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt message")
	}
	if err := msgpack.Unmarshal(sig[keys.SignOverhead:], v); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal message")
	}
	return sig, nil
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

// NewMessageForChannelInvites ...
func NewMessageForChannelInvites(sender keys.ID, invites []*ChannelInvite) *Message {
	msg := NewMessage(sender)
	msg.ChannelInvites = invites
	return msg
}
