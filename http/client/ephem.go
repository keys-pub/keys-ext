package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

// PutEphemeral ...
func (c *Client) PutEphemeral(ctx context.Context, sender keys.ID, recipient keys.ID, b []byte, genCode bool) (*api.EphemResponse, error) {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return nil, err
	}
	if senderKey == nil {
		return nil, keys.NewErrNotFound(sender.String())
	}

	sp := saltpack.NewSaltpack(c.ks)
	encrypted, err := sp.Signcrypt(b, senderKey, recipient, sender)
	if err != nil {
		return nil, err
	}
	path := keys.Path("ephem", senderKey.ID(), recipient)
	vals := url.Values{}
	if genCode {
		vals.Set("code", "1")
	}

	doc, err := c.putDocument(ctx, path, vals, senderKey, bytes.NewReader(encrypted))
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	var resp api.EphemResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetInvite(ctx context.Context, sender keys.ID, code string) (*api.InviteResponse, error) {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/invite?code=%s", url.QueryEscape(code))
	vals := url.Values{}
	doc, err := c.getDocument(ctx, path, vals, senderKey)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	var resp api.InviteResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetEphemeral(ctx context.Context, sender keys.ID, recipient keys.ID) ([]byte, error) {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return nil, err
	}
	path := keys.Path("ephem", sender, recipient)
	vals := url.Values{}
	doc, err := c.getDocument(ctx, path, vals, senderKey)
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
