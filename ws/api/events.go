package api

import (
	"github.com/keys-pub/keys"
	"github.com/vmihailenco/msgpack/v4"
)

// EventPubSub is the pub/sub name for events.
const EventPubSub = "e"

// Event to client.
// JSON is used for websocket clients.
type Event struct {
	Channel keys.ID `json:"channel,omitempty" msgpack:"c,omitempty"`
	User    keys.ID `json:"user,omitempty" msgpack:"u,omitempty"`
	Index   int64   `json:"idx,omitempty" msgpack:"i,omitempty"`
	Token   string  `json:"token,omitempty" msgpack:"t,omitempty"`
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
