package client

import (
	"bytes"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
)

// PutSnap sets data.
func (c *Client) PutSnap(key *keys.EdX25519Key, data []byte) error {
	sp := saltpack.NewSaltpack(c.ks)
	encrypted, err := sp.Signcrypt(data, key, key.ID())
	if err != nil {
		return err
	}
	path := keys.Path("snap", key.ID())
	if _, err := c.put(path, url.Values{}, key, bytes.NewReader(encrypted)); err != nil {
		return err
	}
	return nil
}

// Snap gets data.
func (c *Client) Snap(key *keys.EdX25519Key) ([]byte, error) {
	path := keys.Path("snap", key.ID())

	doc, err := c.getDocument(path, nil, key)
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
	if pk.ID() != key.ID() {
		return nil, errors.Errorf("invalid snap public key")
	}
	return decrypted, nil
}

// DeleteSnap removes data.
func (c *Client) DeleteSnap(key *keys.EdX25519Key) error {
	path := keys.Path("snap", key.ID())
	if _, err := c.delete(path, url.Values{}, key); err != nil {
		return err
	}
	return nil
}
