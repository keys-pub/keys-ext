package client

import (
	"bytes"
	"io/ioutil"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Share data to recipient.
func (c *Client) Share(recipient keys.PublicKey, key keys.Key, data []byte) (string, error) {
	encrypted, err := c.cp.Seal(data, key, recipient)
	if err != nil {
		return "", err
	}

	path := keys.Path("share", recipient.ID(), key.ID())
	resp, err := c.put(path, url.Values{}, key, bytes.NewReader(encrypted))
	if err != nil {
		return "", err
	}
	url := resp.Request.URL
	return url.Scheme + "://" + url.Host + url.Path, nil
}

// Shared returns shared data.
func (c *Client) Shared(recipient keys.Key, kid keys.ID) ([]byte, error) {
	path := keys.Path("share", recipient.ID(), kid)
	resp, err := c.getResponse(path, url.Values{}, recipient)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	decrypted, sender, err := c.cp.Open(data)
	if err != nil {
		return nil, err
	}
	if sender != kid {
		return nil, errors.Errorf("invalid signer id")
	}
	return decrypted, nil
}

// DeleteShare removes a share.
func (c *Client) DeleteShare(recipient keys.PublicKey, key keys.Key) error {
	path := keys.Path("share", recipient.ID(), key.ID())
	_, err := c.delete(path, url.Values{}, key)
	return err
}
