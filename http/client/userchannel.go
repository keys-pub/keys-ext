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

// UserChannels lists channels for user.
func (c *Client) UserChannels(ctx context.Context, user *keys.EdX25519Key) ([]*api.Channel, error) {
	path := dstore.Path("user", user.ID(), "channels")
	params := url.Values{}
	resp, err := c.get(ctx, path, params, http.Authorization(user))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var out api.UserChannelsResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return out.Channels, nil
}

// UserChannelInvites returns all channel invites.
func (c *Client) UserChannelInvites(ctx context.Context, user *keys.EdX25519Key) ([]*api.ChannelInvite, error) {
	path := dstore.Path("user", user.ID(), "invites")
	params := url.Values{}
	resp, err := c.get(ctx, path, params, http.Authorization(user))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var out api.UserChannelInvitesResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return out.Invites, nil
}

// UserChannelInvite returns an invite for user and channel (if one exists).
func (c *Client) UserChannelInvite(ctx context.Context, user *keys.EdX25519Key, channel keys.ID) (*api.ChannelInvite, error) {
	path := dstore.Path("user", user.ID(), "invite", channel)
	params := url.Values{}
	resp, err := c.get(ctx, path, params, http.Authorization(user))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var out api.UserChannelInviteResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return out.Invite, nil
}

// ChannelInviteAccept accepts channel invite.
func (c *Client) ChannelInviteAccept(ctx context.Context, user *keys.EdX25519Key, channel *keys.EdX25519Key) error {
	path := dstore.Path("user", user.ID(), "invite", channel.ID(), "accept")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", user),
		http.NewAuthKey("Authorization-Channel", channel),
	)
	params := url.Values{}
	if _, err := c.post(ctx, path, params, nil, "", auth); err != nil {
		return err
	}
	return nil
}
