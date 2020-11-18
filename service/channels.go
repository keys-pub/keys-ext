package service

import (
	"context"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	kapi "github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/pkg/errors"
)

func (s *service) Channels(ctx context.Context, req *ChannelsRequest) (*ChannelsResponse, error) {
	inbox, err := s.lookup(ctx, req.Inbox, nil)
	if err != nil {
		return nil, err
	}
	if err := s.pullChannels(ctx, inbox); err != nil {
		return nil, err
	}

	channels, err := s.channels(ctx, inbox)
	if err != nil {
		return nil, err
	}

	return &ChannelsResponse{
		Channels: channels,
	}, nil
}

func (s *service) ChannelCreate(ctx context.Context, req *ChannelCreateRequest) (*ChannelCreateResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Errorf("no channel name specified")
	}
	inbox, err := s.lookup(ctx, req.Inbox, nil)
	if err != nil {
		return nil, err
	}
	inboxKey, err := s.edx25519Key(inbox)
	if err != nil {
		return nil, err
	}

	// Create channel key
	channelKey := keys.GenerateEdX25519Key()

	if err := s.client.ChannelCreate(ctx, channelKey, inboxKey); err != nil {
		return nil, err
	}

	if _, _, err := s.vault.SaveKey(kapi.NewKey(channelKey)); err != nil {
		return nil, err
	}

	msg := api.NewMessage()
	msg.ChannelInfo = &api.ChannelInfo{Name: name}
	msg.Timestamp = s.clock.NowMillis()
	if err := s.client.MessageSend(ctx, msg, inboxKey, channelKey); err != nil {
		return nil, err
	}

	return &ChannelCreateResponse{
		Channel: &Channel{
			ID: channelKey.ID().String(),
		},
	}, nil
}

func (s *service) ChannelInvitesCreate(ctx context.Context, req *ChannelInvitesCreateRequest) (*ChannelInvitesCreateResponse, error) {
	cid, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid channel")
	}
	channel, err := s.edx25519Key(cid)
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

	rids := make([]keys.ID, 0, len(req.Recipients))
	for _, invite := range req.Recipients {
		rid, err := s.lookup(ctx, invite, &lookupOpts{Verify: true})
		if err != nil {
			return nil, err
		}
		rids = append(rids, rid)
	}

	if err := s.client.InviteToChannel(ctx, channel, sender, rids...); err != nil {
		return nil, err
	}
	return &ChannelInvitesCreateResponse{}, nil
}

func (s *service) ChannelInviteAccept(ctx context.Context, req *ChannelInviteAcceptRequest) (*ChannelInviteAcceptResponse, error) {
	cid, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid channel")
	}
	iid, err := s.lookup(ctx, req.Inbox, nil)
	if err != nil {
		return nil, err
	}
	inbox, err := s.edx25519Key(iid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get key")
	}

	// Get invite.
	invite, err := s.client.InboxChannelInvite(ctx, inbox, cid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get invite")
	}

	// Save key.
	channel, _, err := invite.Key(inbox)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt channel key")
	}
	if _, _, err := s.vault.SaveKey(kapi.NewKey(channel)); err != nil {
		return nil, err
	}

	// Accept invite.
	if err := s.client.ChannelInviteAccept(ctx, inbox, channel); err != nil {
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

type channelsState struct {
	Channels []*api.Channel `json:"channels"`
}

func (s *service) pullChannels(ctx context.Context, inbox keys.ID) error {
	logger.Infof("Pull channels (%s)...", inbox)

	inboxKey, err := s.edx25519Key(inbox)
	if err != nil {
		return err
	}
	channels, err := s.client.InboxChannels(ctx, inboxKey)
	if err != nil {
		return err
	}

	path := dstore.Path("inbox", inbox, "channels")
	if err := s.db.Set(ctx, path, dstore.From(channelsState{Channels: channels})); err != nil {
		return err
	}

	for _, channel := range channels {
		pullState, err := s.channelPullState(ctx, channel.ID)
		if err != nil {
			return err
		}

		if pullState.Index < channel.Index {
			if err := s.pullMessages(ctx, channel.ID, inbox); err != nil {
				return err
			}
		}
	}

	return nil
}

type channelInfo struct {
	ID              keys.ID `json:"id,omitempty" msgpack:"id,omitempty"`
	Name            string  `json:"name,omitempty" msgpack:"name,omitempty"`
	Description     string  `json:"desc,omitempty" msgpack:"desc,omitempty"`
	Snippet         string  `json:"snippet,omitempty" msgpack:"snippet,omitempty"`
	Timestamp       int64   `json:"ts,omitempty" msgpack:"ts,omitempty"`
	RemoteTimestamp int64   `json:"rts,omitempty" msgpack:"rts,omitempty"`
}

func (s *service) channel(ctx context.Context, channel keys.ID) (*Channel, error) {
	var info channelInfo
	ok, err := s.db.Load(ctx, dstore.Path("channel", channel.ID(), "info"), &info)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &Channel{
			Name: channel.String(),
		}, nil
	}

	return &Channel{
		ID:        info.ID.String(),
		Name:      info.Name,
		Snippet:   info.Snippet,
		UpdatedAt: info.RemoteTimestamp,
	}, nil
}

func (s *service) channels(ctx context.Context, inbox keys.ID) ([]*Channel, error) {
	path := dstore.Path("inbox", inbox, "channels")
	doc, err := s.db.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var state channelsState
	if err := doc.To(&state); err != nil {
		return nil, err
	}

	out := make([]*Channel, 0, len(state.Channels))
	for _, channel := range state.Channels {
		c, err := s.channel(ctx, channel.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	logger.Debugf("Found %d channels in %s", len(out), inbox)
	return out, nil
}
