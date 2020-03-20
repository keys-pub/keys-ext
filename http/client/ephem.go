package client

import (
	"bytes"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
)

// PutEphemeral ...
func (c *Client) PutEphemeral(sender keys.ID, recipient keys.ID, id string, b []byte) error {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return err
	}

	sp := saltpack.NewSaltpack(c.ks)
	encrypted, err := sp.Signcrypt(b, senderKey, recipient, sender)
	if err != nil {
		return err
	}
	path := keys.Path("ephem", senderKey.ID(), recipient, id)
	vals := url.Values{}
	if _, err := c.putDocument(path, vals, senderKey, bytes.NewReader(encrypted)); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetEphemeral(sender keys.ID, recipient keys.ID, id string) ([]byte, error) {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return nil, err
	}
	path := keys.Path("ephem", sender, recipient, id)
	vals := url.Values{}
	doc, err := c.getDocument(path, vals, senderKey)
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
	if pk.ID() != sender && pk.ID() != recipient {
		return nil, errors.Errorf("invalid sender %s", pk.ID())
	}

	return decrypted, nil
}
