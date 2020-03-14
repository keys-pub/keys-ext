package client

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

func (c *Client) PostMessage(sender *keys.EdX25519Key, recipient keys.ID, data []byte) (*api.MessageResponse, error) {
	sp := saltpack.NewSaltpack(c.ks)
	encrypted, err := sp.Encrypt(data, sender.X25519Key(), recipient)
	if err != nil {
		return nil, err
	}

	path := keys.Path("messages", recipient)
	doc, err := c.postDocument(path, url.Values{}, sender, bytes.NewReader(encrypted))
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.Errorf("failed to post message: no response")
	}

	var msg api.MessageResponse
	if err := json.Unmarshal(doc.Data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// Messages ...
func (c *Client) Messages(key *keys.EdX25519Key, version string) (*api.MessagesResponse, error) {
	path := keys.Path("messages", key.ID())

	params := url.Values{}
	params.Add("include", "md")
	params.Add("version", version)

	// TODO: What if we hit limit, we won't have all the messages

	doc, err := c.getDocument(path, params, key)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	var resp api.MessagesResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, err
	}

	// Decrypt messages
	sp := saltpack.NewSaltpack(c.ks)
	for _, msg := range resp.Messages {
		decrypted, pk, err := sp.Decrypt(msg.Data)
		if err != nil {
			return nil, err
		}
		if pk.ID() != key.X25519Key().ID() {
			return nil, errors.Errorf("invalid message kid")
		}
		msg.Data = decrypted
	}
	return &resp, nil
}
