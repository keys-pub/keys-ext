package client

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/encoding"
)

// GenerateToken generates a token.
func GenerateToken() string {
	return encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
}

// Follow recipient, sharing your drop token.
func (c *Client) Follow(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID, token string) error {
	params := url.Values{}
	params.Set("token", token)
	path := dstore.Path("follow", sender.ID(), recipient)
	req := request{
		Method: "PUT",
		Path:   path,
		Body:   []byte(params.Encode()),
		Key:    sender,
	}
	if _, err := c.req(ctx, req); err != nil {
		return err
	}
	return nil
}

// Unfollow recipient.
func (c *Client) Unfollow(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID) error {
	path := dstore.Path("follow", sender.ID(), recipient)
	req := request{
		Method: "DELETE",
		Path:   path,
		Key:    sender,
	}
	if _, err := c.req(ctx, req); err != nil {
		return err
	}
	return nil
}

// Follows lists follows.
func (c *Client) Follows(ctx context.Context, key *keys.EdX25519Key) ([]*api.Follow, error) {
	path := dstore.Path("follows", key.ID())

	resp, err := c.req(ctx, request{Method: "GET", Path: path, Key: key})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	var out api.FollowsResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}

	return out.Follows, nil
}

// FollowedBy looks up follow.
func (c *Client) FollowedBy(ctx context.Context, sender keys.ID, recipient *keys.EdX25519Key) (*api.Follow, error) {
	path := dstore.Path("follow", sender.ID(), recipient)
	req := request{
		Method: "GET",
		Path:   path,
		Key:    recipient,
	}
	resp, err := c.req(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	var out api.FollowResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}

	return out.Follow, nil
}
