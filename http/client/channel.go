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
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// ChannelCreate creates a channel.
func (c *Client) ChannelCreate(ctx context.Context, channel *keys.EdX25519Key, member *keys.EdX25519Key) error {
	path := dstore.Path("channel", channel.ID())
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", member),
		http.NewAuthKey("Authorization-Channel", channel),
	)
	params := url.Values{}
	if _, err := c.put(ctx, path, params, nil, "", auth); err != nil {
		return err
	}
	return nil
}

// ChannelInfoSet sets channel info.
func (c *Client) ChannelInfoSet(ctx context.Context, channel *keys.EdX25519Key, member *keys.EdX25519Key, info *api.ChannelInfo) error {
	if info.Channel != channel.ID() {
		return errors.Errorf("channel info invalid")
	}
	path := dstore.Path("channel", channel.ID(), "info")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", member),
		http.NewAuthKey("Authorization-Channel", channel),
	)
	params := url.Values{}
	b, err := msgpack.Marshal(info)
	if err != nil {
		return err
	}
	encrypted, err := saltpack.Signcrypt(b, false, member, channel.ID())
	if err != nil {
		return err
	}
	contentHash := http.ContentHash(encrypted)
	if _, err := c.put(ctx, path, params, bytes.NewReader(encrypted), contentHash, auth); err != nil {
		return err
	}
	return nil
}

// ChannelInfo gets channel info.
func (c *Client) ChannelInfo(ctx context.Context, channel *keys.EdX25519Key, member *keys.EdX25519Key) (*api.ChannelInfo, error) {
	path := dstore.Path("channel", channel.ID(), "info")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", member),
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
	b, pk, err := saltpack.SigncryptOpen(resp.Data, false, saltpack.NewKeyring(channel))
	if err != nil {
		return nil, err
	}
	var info api.ChannelInfo
	if err := msgpack.Unmarshal(b, &info); err != nil {
		return nil, err
	}
	if pk != nil {
		info.Sender = pk.ID()
	}
	return &info, nil
}

// InviteToChannel invites a recipient to a channel from an existing member.
func (c *Client) InviteToChannel(ctx context.Context, channel *keys.EdX25519Key, member *keys.EdX25519Key, recipient keys.ID) error {
	path := dstore.Path("channel", channel.ID(), "invite")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", member),
		http.NewAuthKey("Authorization-Channel", channel),
	)

	encryptedKey, err := kapi.EncryptKey(kapi.NewKey(channel), member, recipient)
	if err != nil {
		return err
	}

	invite := &api.ChannelInvite{
		Channel:      channel.ID(),
		Recipient:    recipient,
		Sender:       member.ID(),
		EncryptedKey: encryptedKey,
	}

	b, err := json.Marshal(invite)
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
func (c *Client) ChannelInvites(ctx context.Context, channel *keys.EdX25519Key, member *keys.EdX25519Key) ([]*api.ChannelInvite, error) {
	path := dstore.Path("channel", channel.ID(), "invites")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", member),
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

// ChannelMembers returns channel members.
func (c *Client) ChannelMembers(ctx context.Context, channel *keys.EdX25519Key, member *keys.EdX25519Key) ([]*api.ChannelMember, error) {
	path := dstore.Path("channel", channel.ID(), "members")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", member),
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
	var out api.ChannelMembersResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return out.Members, nil
}
