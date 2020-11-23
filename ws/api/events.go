package api

import (
	"github.com/keys-pub/keys"
	"github.com/vmihailenco/msgpack/v4"
)

// EventPubSub is the pub/sub name for events.
const EventPubSub = "e"

// EventType is the type of event.
type EventType string

// Event types.
const (
	// HelloEventType is sent to client after the connect.
	HelloEventType EventType = "hello"

	// ChannelCreateEventType if channel was created.
	ChannelCreatedEventType EventType = "ch-new"

	// ChannelMessageEventType if channel has a new message.
	ChannelMessageEventType EventType = "ch-msg"
)

// Event to client.
// JSON is used for websocket clients.
type Event struct {
	Type EventType `json:"type,omitempty"`

	User keys.ID `json:"user,omitempty"`

	Channel keys.ID `json:"channel,omitempty"`
	Index   int64   `json:"idx,omitempty"`
}

// PubSubEvent is for pub/sub (server to server) events (using msgpack).
type PubSubEvent struct {
	Type EventType `json:"type,omitempty" msgpack:"t,omitempty"`

	Channel keys.ID `json:"channel,omitempty" msgpack:"c,omitempty"`
	User    keys.ID `json:"user,omitempty" msgpack:"u,omitempty"`

	Recipients []keys.ID `json:"recipients,omitempty" msgpack:"r,omitempty"`
	Index      int64     `json:"index,omitempty" msgpack:"i,omitempty"`
}

// Encrypt value into data (msgpack).
func Encrypt(i interface{}, secretKey *[32]byte) ([]byte, error) {
	b, err := msgpack.Marshal(i)
	if err != nil {
		return nil, err
	}
	return keys.SecretBoxSeal(b, secretKey), nil
}

// Decrypt data into value (msgpack).
func Decrypt(b []byte, v interface{}, secretKey *[32]byte) error {
	decrypted, err := keys.SecretBoxOpen(b, secretKey)
	if err != nil {
		return err
	}
	if err := msgpack.Unmarshal(decrypted, v); err != nil {
		return err
	}
	return nil
}
