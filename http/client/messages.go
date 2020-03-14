package client

import (
	"bytes"
	"encoding/json"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

type Message struct {
	ID   string
	Data []byte

	CreatedAt time.Time
	UpdatedAt time.Time
}

// SendMessage posts an encrypted message.
func (c *Client) SendMessage(sender *keys.EdX25519Key, recipient keys.ID, b []byte) (*Message, error) {
	sp := saltpack.NewSaltpack(c.ks)
	encrypted, err := sp.Signcrypt(b, sender, recipient)
	if err != nil {
		return nil, err
	}
	return c.postMessage(sender, recipient, encrypted)
}

func (c *Client) postMessage(sender *keys.EdX25519Key, recipient keys.ID, b []byte) (*Message, error) {
	path := keys.Path("messages", recipient)
	doc, err := c.postDocument(path, url.Values{}, sender, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.Errorf("failed to post message: no response")
	}

	var msg api.MessageResponse
	if err := json.Unmarshal(doc.Data, &msg); err != nil {
		return nil, err
	}
	// TODO: CreatedAt, UpdatedAt
	return &Message{
		ID:   msg.ID,
		Data: b,
	}, nil
}

// Messages returns encrypted messages.
// To decrypt a message, use Client#DecryptMessage.
func (c *Client) Messages(key *keys.EdX25519Key, version string) ([]*Message, string, error) {
	path := keys.Path("messages", key.ID())

	params := url.Values{}
	params.Add("include", "md")
	params.Add("version", version)

	// TODO: What if we hit limit, we won't have all the messages

	doc, err := c.getDocument(path, params, key)
	if err != nil {
		return nil, "", err
	}
	if doc == nil {
		return nil, "", nil
	}

	var resp api.MessagesResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, "", err
	}

	msgs := make([]*Message, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		msgs = append(msgs, &Message{
			ID:        msg.ID,
			Data:      msg.Data,
			CreatedAt: resp.MetadataFor(msg).CreatedAt,
			UpdatedAt: resp.MetadataFor(msg).UpdatedAt,
		})
	}

	return msgs, resp.Version, nil
}

func (c *Client) DecryptMessage(key *keys.EdX25519Key, msg *Message) ([]byte, keys.ID, error) {
	sp := saltpack.NewSaltpack(c.ks)
	decrypted, pk, err := sp.SigncryptOpen(msg.Data)
	if err != nil {
		return nil, "", err
	}
	if pk.ID() != key.ID() {
		return nil, "", errors.Errorf("invalid message kid")
	}
	return decrypted, pk.ID(), nil
}
