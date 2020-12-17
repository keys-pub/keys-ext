package api

import (
	"github.com/keys-pub/keys"
)

// EventPubSub is the pub/sub name for events.
const EventPubSub = "e"

// Event to client.
// JSON is used for websocket clients.
type Event struct {
	Channel keys.ID `json:"channel,omitempty" msgpack:"c,omitempty"`
	Index   int64   `json:"idx,omitempty" msgpack:"i,omitempty"`
}
