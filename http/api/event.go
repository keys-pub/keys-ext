package api

import "github.com/keys-pub/keys/docs/events"

// EventsResponse ...
type EventsResponse struct {
	Events []*events.Event `json:"events"`
	Index  int64           `json:"idx"`
}

// Data for request body.
type Data struct {
	Data []byte `json:"data"`
}
