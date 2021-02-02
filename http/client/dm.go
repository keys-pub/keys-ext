package client

import (
	"context"
	"encoding/json"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
)

// DirectMessageSend an encrypted message.
func (c *Client) DirectMessageSend(ctx context.Context, message *api.Message, sender *keys.EdX25519Key, recipient keys.ID) error {
	encrypted, err := message.Encrypt(sender, recipient)
	if err != nil {
		return err
	}
	path := dstore.Path("dm", sender.ID(), recipient)
	req := &Request{
		Method: "POST",
		Path:   path,
		Body:   encrypted,
		Key:    sender,
	}
	if _, err := c.Request(ctx, req); err != nil {
		return err
	}
	return nil
}

// DirectMessages returns direct messages.
// If truncated, there are more results if you call again with the new index.
// To decrypt to api.Message, use api.DecryptMessageFromEvent.
func (c *Client) DirectMessages(ctx context.Context, key *keys.EdX25519Key, opts *MessagesOpts) (*api.Events, error) {
	path := dstore.Path("dm", key.ID())
	return c.events(ctx, path, key, opts)
}

// DirectToken ...
func (c *Client) DirectToken(ctx context.Context, recipient *keys.EdX25519Key) (*api.DirectToken, error) {
	path := dstore.Path("/dm/token", recipient.ID())
	req := &Request{
		Method: "GET",
		Path:   path,
		Key:    recipient,
	}
	resp, err := c.Request(ctx, req)
	if err != nil {
		return nil, err
	}
	var out api.DirectToken
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
