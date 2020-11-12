package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	kapi "github.com/keys-pub/keys/api"
	"github.com/pkg/errors"
)

func (s *service) Channels(ctx context.Context, req *ChannelsRequest) (*ChannelsResponse, error) {
	mid, err := keys.ParseID(req.Member)
	if err != nil {
		return nil, err
	}
	member, err := s.edx25519Key(mid)
	if err != nil {
		return nil, err
	}
	channels, err := s.client.InboxChannels(ctx, member)
	if err != nil {
		return nil, err
	}
	return &ChannelsResponse{
		Channels: s.channelsToRPC(channels),
	}, nil
}

func (s *service) ChannelCreate(ctx context.Context, req *ChannelCreateRequest) (*ChannelCreateResponse, error) {
	mid, err := keys.ParseID(req.Member)
	if err != nil {
		return nil, err
	}
	member, err := s.edx25519Key(mid)
	if err != nil {
		return nil, err
	}

	channel := keys.GenerateEdX25519Key()

	if _, _, err := s.vault.SaveKey(kapi.NewKey(channel)); err != nil {
		return nil, err
	}

	if err := s.client.ChannelCreate(ctx, channel, member); err != nil {
		return nil, err
	}

	return &ChannelCreateResponse{
		Channel: &Channel{
			ID: channel.ID().String(),
		},
	}, nil
}

func (s *service) ChannelInviteCreate(ctx context.Context, req *ChannelInviteCreateRequest) (*ChannelInviteCreateResponse, error) {
	kid, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid channel")
	}
	channel, err := s.edx25519Key(kid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get channel key")
	}
	sid, err := s.lookup(ctx, req.Sender, nil)
	if err != nil {
		return nil, err
	}
	sender, err := s.edx25519Key(sid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get sender key")
	}
	rid, err := s.lookup(ctx, req.Recipient, nil)
	if err != nil {
		return nil, err
	}
	if err := s.client.InviteToChannel(ctx, channel, sender, rid); err != nil {
		return nil, err
	}
	return &ChannelInviteCreateResponse{}, nil
}

func (s *service) ChannelInviteAccept(ctx context.Context, req *ChannelInviteAcceptRequest) (*ChannelInviteAcceptResponse, error) {
	cid, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid channel")
	}
	mid, err := s.lookup(ctx, req.Member, nil)
	if err != nil {
		return nil, err
	}
	member, err := s.edx25519Key(mid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get member key")
	}

	// Get invite.
	invite, err := s.client.InboxChannelInvite(ctx, member, cid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get invite")
	}

	// Save key.
	channel, _, err := invite.Key(member)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt channel key")
	}
	if _, _, err := s.vault.SaveKey(kapi.NewKey(channel)); err != nil {
		return nil, err
	}

	// Accept invite.
	if err := s.client.ChannelInviteAccept(ctx, member, channel); err != nil {
		return nil, err
	}

	return &ChannelInviteAcceptResponse{}, nil
}

func (s *service) channelsToRPC(cs []*api.Channel) []*Channel {
	out := make([]*Channel, 0, len(cs))
	for _, c := range cs {
		out = append(out, s.channelToRPC(c))
	}
	return out
}

func (s *service) channelToRPC(c *api.Channel) *Channel {
	return &Channel{
		ID: c.ID.String(),
		// Name: c.Name,
	}
}
