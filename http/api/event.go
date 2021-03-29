package api

import (
	"github.com/keys-pub/keys/dstore/events"
)

// EventsResponse ...
type EventsResponse struct {
	Events []*events.Event `json:"events" msgpack:"events"`
	Index  int64           `json:"idx" msgpack:"idx"`
}

// Data for request body.
type Data struct {
	Data []byte `json:"data" msgpack:"dat"`
}

// Events ...
type Events struct {
	Events    []*events.Event `json:"events"`
	Index     int64           `json:"idx"`
	Truncated bool            `json:"truncated,omitempty"`
}
