package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
)

func (c *Client) CreateInvite(ctx context.Context, sender keys.ID, recipient keys.ID) (*api.CreateInviteResponse, error) {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return nil, err
	}
	path := keys.Path("invite", sender, recipient)
	vals := url.Values{}
	doc, err := c.postDocument(ctx, path, vals, senderKey, nil)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	var resp api.CreateInviteResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Invite(ctx context.Context, sender keys.ID, code string) (*api.InviteResponse, error) {
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
