package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
)

// Channels lists channels for user.
func (c *Client) Channels(ctx context.Context, user *keys.EdX25519Key) ([]*api.Channel, error) {
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

// ChannelJoin joins a channel.
func (c *Client) ChannelJoin(ctx context.Context, user *keys.EdX25519Key, channel *keys.EdX25519Key) (*api.Message, error) {
	path := dstore.Path("user", user.ID(), "channel", channel.ID())
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", user),
		http.NewAuthKey("Authorization-Channel", channel),
	)

	// Join message
	msg := api.NewMessageForChannelJoin(user.ID(), user.ID())
	msgEncrypted, err := EncryptMessage(msg, user, channel.ID())
	if err != nil {
		return nil, err
	}
	req := api.ChannelJoinRequest{
		Message: msgEncrypted,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if _, err := c.put(ctx, path, params, bytes.NewReader(body), http.ContentHash(body), auth); err != nil {
		return nil, err
	}
	return msg, nil
}

// ChannelLeave leaves a channel.
func (c *Client) ChannelLeave(ctx context.Context, user *keys.EdX25519Key, channel keys.ID) (*api.Message, error) {
	path := dstore.Path("user", user.ID(), "channel", channel)
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", user),
	)

	// Leave message
	msg := api.NewMessageForChannelLeave(user.ID(), user.ID())
	msgEncrypted, err := EncryptMessage(msg, user, channel.ID())
	if err != nil {
		return nil, err
	}
	req := api.ChannelLeaveRequest{
		Message: msgEncrypted,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if _, err := c.delete(ctx, path, params, bytes.NewReader(body), http.ContentHash(body), auth); err != nil {
		return nil, err
	}
	return msg, nil
}
