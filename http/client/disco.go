package client

import (
	"bytes"
	"context"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

type DiscoType string

const (
	Offer  DiscoType = "offer"
	Answer DiscoType = "answer"
)

// PutDisco puts a discovery offer or answer.
func (c *Client) PutDisco(ctx context.Context, sender keys.ID, recipient keys.ID, typ DiscoType, data string, expire time.Duration) error {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return err
	}
	if senderKey == nil {
		return keys.NewErrNotFound(sender.String())
	}
	recipientKey, err := keys.NewX25519PublicKeyFromID(recipient)
	if err != nil {
		return err
	}
	if expire == time.Duration(0) {
		return errors.Errorf("no expire specified")
	}

	encrypted := keys.BoxSeal([]byte(data), recipientKey, senderKey.X25519Key())

	path := keys.Path("disco", senderKey.ID(), recipient, string(typ))
	vals := url.Values{}
	vals.Set("expire", expire.String())
	if _, err := c.putDocument(ctx, path, vals, senderKey, bytes.NewReader(encrypted)); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetDisco(ctx context.Context, sender keys.ID, recipient keys.ID, typ DiscoType) (string, error) {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return "", err
	}
	recipientKey, err := keys.NewX25519PublicKeyFromID(recipient)
	if err != nil {
		return "", err
	}

	path := keys.Path("disco", sender, recipient, string(typ))
	vals := url.Values{}
	doc, err := c.getDocument(ctx, path, vals, senderKey)
	if err != nil {
		return "", err
	}
	if doc == nil {
		return "", nil
	}

	decrypted, err := keys.BoxOpen(doc.Data, recipientKey, senderKey.X25519Key())
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func (c *Client) DeleteDisco(ctx context.Context, sender keys.ID, recipient keys.ID) error {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return err
	}
	if senderKey == nil {
		return keys.NewErrNotFound(sender.String())
	}

	path := keys.Path("disco", senderKey.ID(), recipient)
	vals := url.Values{}
	if _, err := c.delete(ctx, path, vals, senderKey); err != nil {
		return err
	}
	return nil
}
