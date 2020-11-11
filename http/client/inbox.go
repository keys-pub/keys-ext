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
func (c *Client) InboxChannels(ctx context.Context, key *keys.EdX25519Key) ([]*api.Channel, error) {
	path := dstore.Path("inbox", key.ID(), "channels")
	params := url.Values{}
	resp, err := c.get(ctx, path, params, http.Authorization(key))
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
func (c *Client) InboxChannelInvites(ctx context.Context, key *keys.EdX25519Key) ([]*api.ChannelInvite, error) {
	path := dstore.Path("inbox", key.ID(), "invites")
	params := url.Values{}
	resp, err := c.get(ctx, path, params, http.Authorization(key))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var out api.ChannelInvitesResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return out.Invites, nil
}

// ChannelInviteAccept accepts channel invite.
func (c *Client) ChannelInviteAccept(ctx context.Context, recipient *keys.EdX25519Key, channel *keys.EdX25519Key) error {
	path := dstore.Path("inbox", recipient.ID(), "invite", channel.ID(), "accept")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", recipient),
		http.NewAuthKey("Authorization-Channel", channel),
	)
	params := url.Values{}
	if _, err := c.post(ctx, path, params, nil, "", auth); err != nil {
		return err
	}
	return nil
}
