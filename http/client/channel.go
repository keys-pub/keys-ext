package client

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/pkg/errors"
)

// ChannelCreated ...
type ChannelCreated struct {
	Channel *api.Channel
	Message *api.Message
}

// ChannelCreate creates a channel.
func (c *Client) ChannelCreate(ctx context.Context, channel *keys.EdX25519Key, user *keys.EdX25519Key, info *api.ChannelInfo) (*ChannelCreated, error) {
	path := dstore.Path("channel", channel.ID())

	var msg *api.Message
	var body []byte
	if info != nil {
		msg = api.NewMessageForChannelInfo(user.ID(), info)
		msgEncrypted, err := msg.Encrypt(user, channel.ID())
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
	resp, err := c.Request(ctx, &Request{Method: "PUT", Path: path, Params: params, Body: body, Key: channel})
	if err != nil {
		return nil, err
	}
	var out api.ChannelCreateResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return &ChannelCreated{
		Channel: out.Channel,
		Message: msg,
	}, nil
}

// ChannelsStatus lists channel status.
func (c *Client) ChannelsStatus(ctx context.Context, channelTokens ...*api.ChannelToken) ([]*api.ChannelStatus, error) {
	statusReq := api.ChannelsStatusRequest{
		Channels: map[keys.ID]string{},
	}
	for _, ct := range channelTokens {
		statusReq.Channels[ct.Channel] = ct.Token
	}

	body, err := json.Marshal(statusReq)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	resp, err := c.Request(ctx, &Request{Method: "POST", Path: "/channels/status", Params: params, Body: body})
	if err != nil {
		return nil, err
	}

	var out api.ChannelsStatusResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	sort.Slice(out.Channels, func(i, j int) bool {
		return out.Channels[i].Timestamp > out.Channels[j].Timestamp
	})
	return out.Channels, nil
}

// InviteToChannel sends a direct message containing a channel key.
func (c *Client) InviteToChannel(ctx context.Context, invite *api.ChannelInvite, sender *keys.EdX25519Key) error {
	if sender.ID() != invite.Sender {
		return errors.Errorf("invite sender mismatch")
	}
	msg := api.NewMessageForChannelInvites(invite.Sender, []*api.ChannelInvite{invite})
	if err := c.DirectMessageSend(ctx, msg, sender, invite.Recipient); err != nil {
		return err
	}
	return nil
}

// ChannelInvites ..
type ChannelInvites struct {
	Invites   []*api.ChannelInvite
	Index     int64
	Truncated bool
}

// ChannelInvites lists channel invites from directs.
func (c *Client) ChannelInvites(ctx context.Context, recipient *keys.EdX25519Key, opts *MessagesOpts) (*ChannelInvites, error) {
	directs, err := c.DirectMessages(ctx, recipient, opts)
	if err != nil {
		return nil, err
	}
	msgs, err := directs.Decrypt(recipient)
	if err != nil {
		return nil, err
	}
	invites := []*api.ChannelInvite{}
	for _, msg := range msgs {
		if msg.ChannelInvites != nil {
			invites = append(invites, msg.ChannelInvites...)
		}
	}
	return &ChannelInvites{
		Invites:   invites,
		Index:     directs.Index,
		Truncated: directs.Truncated,
	}, nil
}
