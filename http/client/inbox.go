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

// InboxChannels lists channels in inbox.
func (c *Client) InboxChannels(ctx context.Context, inbox *keys.EdX25519Key) ([]*api.Channel, error) {
	path := dstore.Path("inbox", inbox.ID(), "channels")
	params := url.Values{}
	resp, err := c.get(ctx, path, params, http.Authorization(inbox))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var out api.InboxChannelsResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return out.Channels, nil
}

// InboxChannelInvites ...
func (c *Client) InboxChannelInvites(ctx context.Context, inbox *keys.EdX25519Key) ([]*api.ChannelInvite, error) {
	path := dstore.Path("inbox", inbox.ID(), "invites")
	params := url.Values{}
	resp, err := c.get(ctx, path, params, http.Authorization(inbox))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var out api.InboxChannelInvitesResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return out.Invites, nil
}

// InboxChannelInvite ...
func (c *Client) InboxChannelInvite(ctx context.Context, inbox *keys.EdX25519Key, channel keys.ID) (*api.ChannelInvite, error) {
	path := dstore.Path("inbox", inbox.ID(), "invite", channel)
	params := url.Values{}
	resp, err := c.get(ctx, path, params, http.Authorization(inbox))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var out api.InboxChannelInviteResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return out.Invite, nil
}

// ChannelInviteAccept accepts channel invite.
func (c *Client) ChannelInviteAccept(ctx context.Context, inbox *keys.EdX25519Key, channel *keys.EdX25519Key) error {
	path := dstore.Path("inbox", inbox.ID(), "invite", channel.ID(), "accept")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", inbox),
		http.NewAuthKey("Authorization-Channel", channel),
	)
	params := url.Values{}
	if _, err := c.post(ctx, path, params, nil, "", auth); err != nil {
		return err
	}
	return nil
}
