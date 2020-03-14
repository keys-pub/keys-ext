package client

import (
	"net/url"

	"github.com/keys-pub/keys"
)

// Check user & sigchain associated with edx25519 key.
// The server periodically checks users and sigchains, but this tells the server
// to do it right away.
func (c *Client) Check(key *keys.EdX25519Key) error {
	params := url.Values{}
	_, err := c.post("/check", params, key, nil)
	if err != nil {
		return err
	}
	return nil
}
