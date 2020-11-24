package service

import (
	"context"
	"sort"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	kapi "github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/pkg/errors"
)

func (s *service) Channels(ctx context.Context, req *ChannelsRequest) (*ChannelsResponse, error) {
	user, err := s.lookup(ctx, req.User, nil)
	if err != nil {
		return nil, err
	}
	if err := s.pullChannels(ctx, user); err != nil {
		return nil, err
	}

	channels, err := s.channels(ctx, user)
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
	user, err := s.lookup(ctx, req.User, nil)
	if err != nil {
		return nil, err
	}
	userKey, err := s.edx25519Key(user)
	if err != nil {
		return nil, err
	}

	// Create channel key
	channelKey := keys.GenerateEdX25519Key()

	if err := s.client.ChannelCreate(ctx, channelKey, userKey); err != nil {
		return nil, err
	}

	if _, _, err := s.vault.SaveKey(kapi.NewKey(channelKey)); err != nil {
		return nil, err
	}

	msg := api.NewMessage()
	msg.ChannelInfo = &api.ChannelInfo{Name: name}
	msg.Timestamp = s.clock.NowMillis()
	if err := s.client.MessageSend(ctx, msg, userKey, channelKey); err != nil {
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
	uid, err := s.lookup(ctx, req.User, nil)
	if err != nil {
		return nil, err
	}
	user, err := s.edx25519Key(uid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get key")
	}

	// Get invite.
	invite, err := s.client.UserChannelInvite(ctx, user, cid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get invite")
	}

	// Save key.
	channel, _, err := invite.Key(user)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt channel key")
	}
	if _, _, err := s.vault.SaveKey(kapi.NewKey(channel)); err != nil {
		return nil, err
	}

	// Accept invite.
	if err := s.client.ChannelInviteAccept(ctx, user, channel); err != nil {
		return nil, err
	}

	return &ChannelInviteAcceptResponse{}, nil
}

type channelsState struct {
	Channels []*api.Channel `json:"channels"`
}

func (s *service) pullChannels(ctx context.Context, user keys.ID) error {
	logger.Infof("Pull channels (%s)...", user)

	userKey, err := s.edx25519Key(user)
	if err != nil {
		return err
	}
	channels, err := s.client.UserChannels(ctx, userKey)
	if err != nil {
		return err
	}
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Timestamp > channels[j].Timestamp
	})

	path := dstore.Path("users", user, "channels")
	if err := s.db.Set(ctx, path, dstore.From(channelsState{Channels: channels})); err != nil {
		return err
	}

	// TODO: Pull channels in a single (bulk) call
	for _, channel := range channels {
		pullState, err := s.channelPullState(ctx, channel.ID)
		if err != nil {
			return err
		}

		if pullState.Index < channel.Index {
			if err := s.pullMessages(ctx, channel.ID, user); err != nil {
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
	Index           int64   `json:"index,omitempty" msgpack:"index,omitempty"`
	Timestamp       int64   `json:"ts,omitempty" msgpack:"ts,omitempty"`
	RemoteTimestamp int64   `json:"rts,omitempty" msgpack:"rts,omitempty"`
}

func (s *service) channel(ctx context.Context, channel keys.ID) (*Channel, error) {
	// channelInfo is set during pullMessages
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
		Index:     info.Index,
	}, nil
}

func (s *service) channels(ctx context.Context, users keys.ID) ([]*Channel, error) {
	path := dstore.Path("users", users, "channels")
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
	logger.Debugf("Found %d channels in %s", len(out), users)
	return out, nil
}
