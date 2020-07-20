package api

import "github.com/keys-pub/keys/docs/events"

// MessagesResponse ...
type MessagesResponse struct {
	Messages []*events.Event `json:"msgs"`
	Index    int64           `json:"idx"`
}
