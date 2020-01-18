package client

import (
	"net/url"

	"github.com/keys-pub/keys"
)

// Check ...
func (c *Client) Check(key *keys.SignKey) error {
	params := url.Values{}
	_, err := c.post("/check", params, key, nil)
	if err != nil {
		return err
	}
	return nil
}
