package client

import (
	"bytes"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
)

// PutEphemeral ...
func (c *Client) PutEphemeral(sender *keys.EdX25519Key, recipient keys.ID, id string, b []byte) error {
	sp := saltpack.NewSaltpack(c.ks)
	encrypted, err := sp.Signcrypt(b, sender, recipient, sender.ID())
	if err != nil {
		return err
	}
	path := keys.Path("ephem", sender.ID(), recipient, id)
	vals := url.Values{}
	if _, err := c.putDocument(path, vals, sender, bytes.NewReader(encrypted)); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetEphemeral(sender *keys.EdX25519Key, recipient keys.ID, id string) ([]byte, error) {
	path := keys.Path("ephem", sender.ID(), recipient, id)
	vals := url.Values{}
	doc, err := c.getDocument(path, vals, sender)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	sp := saltpack.NewSaltpack(c.ks)
	decrypted, pk, err := sp.SigncryptOpen(doc.Data)
	if err != nil {
		return nil, err
	}
	if pk.ID() != sender.ID() && pk.ID() != recipient {
		return nil, errors.Errorf("invalid sender %s", pk.ID())
	}

	return decrypted, nil
}
