package api

import (
	"github.com/keys-pub/keys"
)

// EventPubSub is the pub/sub key/name for events.
const EventPubSub = "e"

// EventType is the type of event.
type EventType string

// Event types.
const (
	// HelloEvent is sent to client after the connect.
	HelloEvent EventType = "hello"
	// ChannelEvent is sent to client if a channel has changed.
	ChannelEvent EventType = "channel"
)

// Event to client.
type Event struct {
	Type EventType `json:"type,omitempty"`

	Channel keys.ID `json:"channel,omitempty"`
	User    keys.ID `json:"user,omitempty"`

	Index int64 `json:"idx,omitempty"`
}

// PubEvent notification sent through Redis pub/sub from server to websockets
// to notify clients of a channel message.
type PubEvent struct {
	Channel keys.ID   `json:"channel,omitempty" msgpack:"c,omitempty"`
	Users   []keys.ID `json:"users,omitempty" msgpack:"u,omitempty"`
	Index   int64     `json:"index,omitempty" msgpack:"i,omitempty"`
}
