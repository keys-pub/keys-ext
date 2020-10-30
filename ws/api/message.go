package api

import (
	"github.com/keys-pub/keys"
)

// MessageType is the type of message.
type MessageType string

// Message types.
const (
	Hello   MessageType = "hello"
	Changed MessageType = "chg"
)

// Message to client.
type Message struct {
	KID  keys.ID     `json:"kid"`
	Type MessageType `json:"type,omitempty"`
}
