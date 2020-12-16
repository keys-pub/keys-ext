package client

import (
	"bytes"
	"context"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
)

// ShareSeal saves a secret on remote with expire.
func (c *Client) ShareSeal(ctx context.Context, key *keys.EdX25519Key, data []byte, expire time.Duration) error {
	encrypted := keys.BoxSeal(data, key.X25519Key().PublicKey(), key.X25519Key())

	path := dstore.Path("share", key.ID())
	vals := url.Values{}
	vals.Set("expire", expire.String())
	if _, err := c.put(ctx, path, vals, bytes.NewReader(encrypted), http.ContentHash(encrypted), key); err != nil {
		return err
	}
	return nil
}

// ShareOpen opens a secret.
func (c *Client) ShareOpen(ctx context.Context, key *keys.EdX25519Key) ([]byte, error) {
	path := dstore.Path("share", key.ID())
	vals := url.Values{}
	resp, err := c.get(ctx, path, vals, key)
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
