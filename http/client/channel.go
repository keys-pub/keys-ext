package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	kapi "github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
)

// ChannelCreate creates a channel.
func (c *Client) ChannelCreate(ctx context.Context, channel *keys.EdX25519Key, user *keys.EdX25519Key) error {
	path := dstore.Path("channel", channel.ID())
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", user),
		http.NewAuthKey("Authorization-Channel", channel),
	)
	params := url.Values{}
	if _, err := c.put(ctx, path, params, nil, "", auth); err != nil {
		return err
	}
	return nil
}

// InviteToChannel invites a recipient to a channel from an existing user.
func (c *Client) InviteToChannel(ctx context.Context, channel *keys.EdX25519Key, user *keys.EdX25519Key, recipients ...keys.ID) error {
	path := dstore.Path("channel", channel.ID(), "invites")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", user),
		http.NewAuthKey("Authorization-Channel", channel),
	)
	invites := make([]*api.ChannelInvite, 0, len(recipients))
	for _, recipient := range recipients {
		encryptedKey, err := kapi.EncryptKey(kapi.NewKey(channel), user, recipient)
		if err != nil {
			return err
		}

		invite := &api.ChannelInvite{
			Channel:      channel.ID(),
			Recipient:    recipient,
			Sender:       user.ID(),
			EncryptedKey: encryptedKey,
		}
		invites = append(invites, invite)
	}

	b, err := json.Marshal(invites)
	if err != nil {
		return err
	}

	params := url.Values{}
	if _, err := c.post(ctx, path, params, bytes.NewReader(b), http.ContentHash(b), auth); err != nil {
		return err
	}
	return nil
}

// ChannelInvites returns all pending invites.
func (c *Client) ChannelInvites(ctx context.Context, channel *keys.EdX25519Key, user *keys.EdX25519Key) ([]*api.ChannelInvite, error) {
	path := dstore.Path("channel", channel.ID(), "invites")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", user),
		http.NewAuthKey("Authorization-Channel", channel),
	)
	params := url.Values{}
	resp, err := c.get(ctx, path, params, auth)
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

// ChannelUsers returns channel users.
func (c *Client) ChannelUsers(ctx context.Context, channel *keys.EdX25519Key, user *keys.EdX25519Key) ([]*api.ChannelUser, error) {
	path := dstore.Path("channel", channel.ID(), "users")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", user),
		http.NewAuthKey("Authorization-Channel", channel),
	)
	params := url.Values{}
	resp, err := c.get(ctx, path, params, auth)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var out api.ChannelUsersResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return out.Users, nil
}
