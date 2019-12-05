package client

import (
	"bytes"
	"encoding/json"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
)

// Item ...
type Item struct {
	// Data is encrypted.
	Data []byte
	// ID for message.
	ID keys.ID
	// CreatedAt ...
	CreatedAt time.Time
}

// Items ...
type Items struct {
	Items   []*Item
	Version string
}

// PutItem ...
func (c *Client) PutItem(sender keys.Key, key keys.Key, id keys.ID, data []byte) (*Item, error) {
	encrypted, err := c.cp.Seal(data, sender, key.PublicKey())
	if err != nil {
		return nil, err
	}
	path := keys.Path("vault", key.ID(), id)
	if _, err := c.put(path, url.Values{}, key, bytes.NewReader(encrypted)); err != nil {
		return nil, err
	}
	return &Item{
		Data: encrypted,
		ID:   id,
	}, nil
}

// Vault ...
func (c *Client) Vault(key keys.Key, version string) (*api.VaultResponse, error) {
	path := keys.Path("vault", key.ID())

	params := url.Values{}
	params.Add("include", "md")
	params.Add("version", version)

	e, err := c.get(path, params, key)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, nil
	}

	var resp api.VaultResponse
	if err := json.Unmarshal(e.Data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
