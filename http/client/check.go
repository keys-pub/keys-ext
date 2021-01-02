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
	if _, err := c.req(ctx, request{Method: "POST", Path: "/check", Params: params, Key: key}); err != nil {
		return err
	}
	return nil
}
