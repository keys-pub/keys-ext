package api

type Event struct {
	Data []byte `json:"data" msgpack:"dat" firestore:"data"`

	// Index for event (read only).
	Index int64 `json:"idx" msgpack:"idx" firestore:"idx"`
	// Timestamp (read only). The time at which the event was created.
	// Firestore sets this to the document create time.
	Timestamp int64 `json:"ts" msgpack:"ts" firestore:"-"`
}

// EventsResponse ...
type EventsResponse struct {
	Events []*Event `json:"events" msgpack:"events"`
	Index  int64    `json:"idx" msgpack:"idx"`
}

// Data for request body.
type Data struct {
	Data []byte `json:"data" msgpack:"dat"`
}

// Events ...
type Events struct {
	Events    []*Event `json:"events"`
	Index     int64    `json:"idx"`
	Truncated bool     `json:"truncated,omitempty"`
}
