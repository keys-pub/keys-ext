package client

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
)

// InviteCodeCreate creates an invite code.
func (c *Client) InviteCodeCreate(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID) (*api.InviteCodeCreateResponse, error) {
	path := dstore.Path("/invite/code", sender.ID(), recipient)
	vals := url.Values{}
	resp, err := c.post(ctx, path, vals, nil, "", http.Authorization(sender))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var out api.InviteCodeCreateResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// InviteCode looks for an invite with code.
func (c *Client) InviteCode(ctx context.Context, sender *keys.EdX25519Key, code string) (*api.InviteCodeResponse, error) {
	path := dstore.Path("/invite/code", url.QueryEscape(code))
	vals := url.Values{}
	resp, err := c.get(ctx, path, vals, http.Authorization(sender))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var out api.InviteCodeResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
