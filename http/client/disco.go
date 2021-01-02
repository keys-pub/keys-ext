package client

import (
	"context"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
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

// DiscoSave puts a discovery offer or answer.
func (c *Client) DiscoSave(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID, typ DiscoType, data string, expire time.Duration) error {
	recipientKey, err := keys.NewX25519PublicKeyFromID(recipient)
	if err != nil {
		return err
	}
	if expire == time.Duration(0) {
		return errors.Errorf("no expire specified")
	}

	encrypted := keys.BoxSeal([]byte(data), recipientKey, sender.X25519Key())

	path := dstore.Path("disco", sender.ID(), recipient, string(typ))
	vals := url.Values{}
	vals.Set("expire", expire.String())
	if _, err := c.req(ctx, request{Method: "PUT", Path: path, Params: vals, Body: encrypted, Key: sender}); err != nil {
		return err
	}
	return nil
}

// Disco gets a discovery address.
func (c *Client) Disco(ctx context.Context, sender keys.ID, recipient *keys.EdX25519Key, typ DiscoType) (string, error) {
	senderKey, err := keys.NewX25519PublicKeyFromID(sender)
	if err != nil {
		return "", err
	}

	path := dstore.Path("disco", sender, recipient, string(typ))
	vals := url.Values{}
	resp, err := c.req(ctx, request{Method: "GET", Path: path, Params: vals, Key: recipient})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", nil
	}

	decrypted, err := keys.BoxOpen(resp.Data, senderKey, recipient.X25519Key())
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// DiscoDelete removes discovery addresses.
func (c *Client) DiscoDelete(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID) error {
	path := dstore.Path("disco", sender.ID(), recipient)
	vals := url.Values{}
	if _, err := c.req(ctx, request{Method: "DELETE", Path: path, Params: vals, Key: sender}); err != nil {
		return err
	}
	return nil
}
