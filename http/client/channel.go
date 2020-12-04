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
func (c *Client) ChannelCreate(ctx context.Context, channel *keys.EdX25519Key, user *keys.EdX25519Key, info *api.ChannelInfo) (*api.Message, error) {
	path := dstore.Path("channel", channel.ID())
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", user),
		http.NewAuthKey("Authorization-Channel", channel),
	)

	var msg *api.Message
	var body []byte
	if info != nil {
		msg = api.NewMessageForChannelInfo(user.ID(), info)
		msgEncrypted, err := EncryptMessage(msg, user, channel.ID())
		if err != nil {
			return nil, err
		}
		req := api.ChannelCreateRequest{
			Message: msgEncrypted,
		}
		b, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		body = b
	}

	params := url.Values{}
	if _, err := c.put(ctx, path, params, bytes.NewReader(body), http.ContentHash(body), auth); err != nil {
		return nil, err
	}
	return msg, nil
}

// InviteToChannel invites a recipient to a channel from an existing user.
func (c *Client) InviteToChannel(ctx context.Context, channel *keys.EdX25519Key, user *keys.EdX25519Key, recipients ...keys.ID) (*api.Message, error) {
	path := dstore.Path("channel", channel.ID(), "invites")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", user),
		http.NewAuthKey("Authorization-Channel", channel),
	)
	invites := make([]*api.ChannelInvite, 0, len(recipients))
	for _, recipient := range recipients {
		encryptedKey, err := kapi.EncryptKey(kapi.NewKey(channel), user, recipient)
		if err != nil {
			return nil, err
		}

		invite := &api.ChannelInvite{
			Channel:      channel.ID(),
			Recipient:    recipient,
			Sender:       user.ID(),
			EncryptedKey: encryptedKey,
		}
		invites = append(invites, invite)
	}

	msg := api.NewMessageForChannelInvites(user.ID(), recipients...)
	msgEncrypted, err := EncryptMessage(msg, user, channel.ID())
	if err != nil {
		return nil, err
	}
	req := api.ChannelInvitesRequest{
		Invites: invites,
		Message: msgEncrypted,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if _, err := c.post(ctx, path, params, bytes.NewReader(body), http.ContentHash(body), auth); err != nil {
		return nil, err
	}
	return msg, nil
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
