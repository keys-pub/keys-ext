package wormhole

import "github.com/keys-pub/keys"

// MessageType describes wormhole message.
type MessageType string

const (
	// Sent for a sent message.
	Sent MessageType = "sent"
	// Pending for a pending message.
	Pending MessageType = "pending"
	// Ack for an acknowledgement.
	Ack MessageType = "ack"
)

// Message for wormhole.
type Message struct {
	ID        string
	Sender    keys.ID
	Recipient keys.ID
	Content   *Content
	Type      MessageType
}

// Content for message.
type Content struct {
	Data []byte
	Type ContentType
}

// ContentType describes the content  type.
type ContentType string

// UTF8Content is UTF8.
const UTF8Content ContentType = "utf8"

// BinaryContent for binary content.
const BinaryContent ContentType = "binary"
