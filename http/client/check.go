package client

import (
	"context"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/http"
)

// Check user & sigchain associated with edx25519 key.
// The server periodically checks users and sigchains, but this tells the server
// to do it right away.
func (c *Client) Check(ctx context.Context, key *keys.EdX25519Key) error {
	params := url.Values{}
	_, err := c.post(ctx, "/check", params, nil, "", http.Authorization(key))
	if err != nil {
		return err
	}
	return nil
}
