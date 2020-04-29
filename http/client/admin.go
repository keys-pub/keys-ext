package client

import (
	"context"
	"net/url"

	"github.com/keys-pub/keys"
)

// AdminCheck performs user & sigchain associated with key by an admin.
// The server periodically checks users and sigchains, but this tells the server
// to do it right away.
func (c *Client) AdminCheck(ctx context.Context, kid keys.ID, admin *keys.EdX25519Key) error {
	params := url.Values{}
	_, err := c.post(ctx, "/admin/check/"+kid.String(), params, admin, nil)
	if err != nil {
		return err
	}
	return nil
}
