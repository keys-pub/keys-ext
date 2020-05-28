package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys-ext/http/api"
)

// CreateInvite writes a sender recipient address (invite).
func (c *Client) CreateInvite(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID) (*api.CreateInviteResponse, error) {
	path := ds.Path("invite", sender.ID(), recipient)
	vals := url.Values{}
	doc, err := c.postDocument(ctx, path, vals, sender, nil)
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

// Invite looks for an invite with code.
func (c *Client) Invite(ctx context.Context, sender *keys.EdX25519Key, code string) (*api.InviteResponse, error) {
	path := fmt.Sprintf("/invite?code=%s", url.QueryEscape(code))
	vals := url.Values{}
	doc, err := c.getDocument(ctx, path, vals, sender)
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
