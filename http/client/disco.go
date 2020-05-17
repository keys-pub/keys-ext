package client

import (
	"bytes"
	"context"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
)

// DiscoType is the type of discovery address.
type DiscoType string

const (
	// Offer initiates.
	Offer DiscoType = "offer"
	// Answer listens.
	Answer DiscoType = "answer"
)

// PutDisco puts a discovery offer or answer.
func (c *Client) PutDisco(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID, typ DiscoType, data string, expire time.Duration) error {
	recipientKey, err := keys.NewX25519PublicKeyFromID(recipient)
	if err != nil {
		return err
	}
	if expire == time.Duration(0) {
		return errors.Errorf("no expire specified")
	}

	encrypted := keys.BoxSeal([]byte(data), recipientKey, sender.X25519Key())

	path := ds.Path("disco", sender.ID(), recipient, string(typ))
	vals := url.Values{}
	vals.Set("expire", expire.String())
	if _, err := c.putDocument(ctx, path, vals, sender, bytes.NewReader(encrypted)); err != nil {
		return err
	}
	return nil
}

// GetDisco gets a discovery address.
func (c *Client) GetDisco(ctx context.Context, sender keys.ID, recipient *keys.EdX25519Key, typ DiscoType) (string, error) {
	senderKey, err := keys.NewX25519PublicKeyFromID(sender)
	if err != nil {
		return "", err
	}

	path := ds.Path("disco", sender, recipient, string(typ))
	vals := url.Values{}
	doc, err := c.getDocument(ctx, path, vals, recipient)
	if err != nil {
		return "", err
	}
	if doc == nil {
		return "", nil
	}

	decrypted, err := keys.BoxOpen(doc.Data, senderKey, recipient.X25519Key())
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// DeleteDisco removes discovery addresses.
func (c *Client) DeleteDisco(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID) error {
	path := ds.Path("disco", sender.ID(), recipient)
	vals := url.Values{}
	if _, err := c.delete(ctx, path, vals, sender); err != nil {
		return err
	}
	return nil
}
