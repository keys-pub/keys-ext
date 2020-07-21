package client

import (
	"bytes"
	"context"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/docs"
)

// ShareSeal saves a secret on remote with expire.
func (c *Client) ShareSeal(ctx context.Context, key *keys.EdX25519Key, data []byte, expire time.Duration) error {
	encrypted := keys.BoxSeal(data, key.X25519Key().PublicKey(), key.X25519Key())
	contentHash := api.ContentHash(encrypted)

	path := docs.Path("share", key.ID())
	vals := url.Values{}
	vals.Set("expire", expire.String())
	if _, err := c.putDocument(ctx, path, vals, key, bytes.NewReader(encrypted), contentHash); err != nil {
		return err
	}
	return nil
}

// ShareOpen opens a secret.
func (c *Client) ShareOpen(ctx context.Context, key *keys.EdX25519Key) ([]byte, error) {
	path := docs.Path("share", key.ID())
	vals := url.Values{}
	doc, err := c.getDocument(ctx, path, vals, key)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	decrypted, err := keys.BoxOpen(doc.Data, key.X25519Key().PublicKey(), key.X25519Key())
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}
