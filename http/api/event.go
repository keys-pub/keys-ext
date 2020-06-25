package api

import "github.com/keys-pub/keys/ds"

// EventsResponse ...
type EventsResponse struct {
	Events []*ds.Event `json:"events"`
	Index  int64       `json:"idx"`
}

// Data for request body.
type Data struct {
	Data []byte `json:"data"`
}
