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
	if _, err := c.put(ctx, path, params, bytes.NewReader(body), http.ContentHash(body), channel); err != nil {
		return nil, err
	}
	return msg, nil
}

// InviteToChannel sends a direct message containing a channel key.
func (c *Client) InviteToChannel(ctx context.Context, channel *keys.EdX25519Key, info *api.ChannelInfo, sender *keys.EdX25519Key, recipient keys.ID) (*api.Message, error) {
	invite := &api.ChannelInvite{
		Channel:   channel.ID(),
		Recipient: recipient,
		Key:       kapi.NewKey(channel),
		Info:      info,
	}

	msg := api.NewMessageForChannelInvite(sender.ID(), invite)
	if err := c.DirectMessageSend(ctx, msg, sender, recipient); err != nil {
		return nil, err
	}
	return msg, nil
}
