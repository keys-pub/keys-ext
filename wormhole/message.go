package wormhole

import "github.com/keys-pub/keys"

type MessageType string

const Sent MessageType = "sent"
const Pending MessageType = "pending"
const Ack MessageType = "ack"

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

type ContentType string

// UTF8Content is utf8.
const UTF8Content ContentType = "utf8"

// BinaryContent for binary content.
const BinaryContent ContentType = "binary"
