package client

import (
	"bytes"
	"encoding/json"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
)

// Message ...
type Message struct {
	// Data is encrypted.
	Data []byte
	// ID for message.
	ID keys.ID
	// CreatedAt ...
	CreatedAt time.Time
}

// Messages ...
type Messages struct {
	Messages []*Message
	Version  string
}

// PutMessage ...
func (c *Client) PutMessage(sender keys.Key, recipient keys.Key, mid keys.ID, data []byte) (*Message, error) {
	encrypted, err := c.cp.Seal(data, sender, recipient.PublicKey())
	if err != nil {
		return nil, err
	}
	path := keys.Path("messages", recipient.ID(), mid)
	if _, err = c.put(path, url.Values{}, recipient, bytes.NewReader(encrypted)); err != nil {
		return nil, err
	}
	return &Message{
		Data: encrypted,
		ID:   mid,
	}, nil
}

// Messages ...
func (c *Client) Messages(recipient keys.Key, version string) (*api.MessagesResponse, error) {
	path := keys.Path("messages", recipient.ID())

	params := url.Values{}
	params.Add("include", "md")
	params.Add("version", version)

	e, err := c.get(path, params, recipient)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, nil
	}

	var resp api.MessagesResponse
	if err := json.Unmarshal(e.Data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
