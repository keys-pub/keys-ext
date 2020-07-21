package client

import (
	"context"
	"net/url"

	"github.com/keys-pub/keys"
)

// Check user & sigchain associated with edx25519 key.
// The server periodically checks users and sigchains, but this tells the server
// to do it right away.
func (c *Client) Check(ctx context.Context, key *keys.EdX25519Key) error {
	params := url.Values{}
	_, err := c.post(ctx, "/check", params, key, nil, "")
	if err != nil {
		return err
	}
	return nil
}
