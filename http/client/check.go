package client

import (
	"net/url"

	"github.com/keys-pub/keys"
)

// Check ...
func (c *Client) Check(key *keys.EdX25519Key) error {
	params := url.Values{}
	_, err := c.post("/check", params, key, nil)
	if err != nil {
		return err
	}
	return nil
}
