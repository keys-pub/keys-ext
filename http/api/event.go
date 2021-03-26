package api

import (
	"github.com/keys-pub/keys"
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
