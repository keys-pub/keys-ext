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
	// Data ...
	Data []byte
	// ID ...
	ID string
	// CreatedAt ...
	CreatedAt time.Time
}

// Messages ...
type Messages struct {
	Messages []*Message
	Version  string
}

// PutMessage ...
func (c *Client) PutMessage(key *keys.EdX25519Key, id string, data []byte) error {
	path := keys.Path("messages", key.ID(), id)
	if _, err := c.put(path, url.Values{}, key, bytes.NewReader(data)); err != nil {
		return err
	}
	return nil
}

// Messages ...
func (c *Client) Messages(key *keys.EdX25519Key, version string) (*api.MessagesResponse, error) {
	path := keys.Path("messages", key.ID())

	params := url.Values{}
	params.Add("include", "md")
	params.Add("version", version)

	// TODO: What if we hit limit, we won't have all the messages

	e, err := c.get(path, params, key)
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
