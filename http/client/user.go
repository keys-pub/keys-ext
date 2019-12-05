package client

import (
	"net/url"

	"github.com/keys-pub/keys"
)

// Check ...
func (c *Client) Check(key keys.Key) error {
	params := url.Values{}
	_, poerr := c.post("/check", params, key, nil)
	if poerr != nil {
		return poerr
	}
	return nil
}
