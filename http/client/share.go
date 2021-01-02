package client

import (
	"context"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
)

// ShareSeal saves a secret on remote with expire.
func (c *Client) ShareSeal(ctx context.Context, key *keys.EdX25519Key, data []byte, expire time.Duration) error {
	encrypted := keys.BoxSeal(data, key.X25519Key().PublicKey(), key.X25519Key())

	path := dstore.Path("share", key.ID())
	vals := url.Values{}
	vals.Set("expire", expire.String())
	if _, err := c.req(ctx, request{Method: "PUT", Path: path, Params: vals, Body: encrypted, Key: key}); err != nil {
		return err
	}
	return nil
}

// ShareOpen opens a secret.
func (c *Client) ShareOpen(ctx context.Context, key *keys.EdX25519Key) ([]byte, error) {
	path := dstore.Path("share", key.ID())
	vals := url.Values{}
	resp, err := c.req(ctx, request{Method: "GET", Path: path, Params: vals, Key: key})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	decrypted, err := keys.BoxOpen(resp.Data, key.X25519Key().PublicKey(), key.X25519Key())
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}
