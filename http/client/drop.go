package client

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
)

// Drop an encrypted message.
func (c *Client) Drop(ctx context.Context, message *api.Message, sender *keys.EdX25519Key, recipient keys.ID) error {
	encrypted, err := message.Encrypt(sender, recipient)
	if err != nil {
		return err
	}
	path := dstore.Path("drop", sender.ID(), recipient)
	req := request{
		Method: "POST",
		Path:   path,
		Body:   encrypted,
		Key:    sender,
	}
	if _, err := c.req(ctx, req); err != nil {
		return err
	}
	return nil
}

// Drops returns encrypted messages.
// If truncated, there are more results if you call again with the new index.
// To decrypt to api.Message, use DecryptMessage.
func (c *Client) Drops(ctx context.Context, key *keys.EdX25519Key, opts *MessagesOpts) (*Messages, error) {
	path := dstore.Path("drop", key.ID())
	return c.messages(ctx, path, key, opts)
}
