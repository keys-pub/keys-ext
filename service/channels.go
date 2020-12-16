package service

import (
	"context"
	"sort"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	kapi "github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/saltpack"
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
	if len(name) > 16 {
		return nil, errors.Errorf("channel name too long (must be < 16)")
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

	info := &api.ChannelInfo{Name: name}
	if _, err := s.client.ChannelCreate(ctx, channelKey, userKey, info); err != nil {
		return nil, err
	}

	if _, _, err := s.vault.SaveKey(kapi.NewKey(channelKey)); err != nil {
		return nil, err
	}

	return &ChannelCreateResponse{
		Channel: &Channel{
			ID: channelKey.ID().String(),
		},
	}, nil
}

func (s *service) ChannelLeave(ctx context.Context, req *ChannelLeaveRequest) (*ChannelLeaveResponse, error) {
	cid, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid channel")
	}
	user, err := s.lookup(ctx, req.User, nil)
	if err != nil {
		return nil, err
	}
	userKey, err := s.edx25519Key(user)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user key")
	}

	msg, err := s.client.ChannelLeave(ctx, userKey, cid)
	if err != nil {
		return nil, err
	}
	message, err := s.messageToRPC(ctx, msg)
	if err != nil {
		return nil, err
	}

	if _, err := s.vault.Delete(cid.String()); err != nil {
		return nil, err
	}

	return &ChannelLeaveResponse{Message: message}, nil
}

func (s *service) ChannelInvitesCreate(ctx context.Context, req *ChannelInvitesCreateRequest) (*ChannelInvitesCreateResponse, error) {
	channel, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid channel")
	}
	channelKey, err := s.edx25519Key(channel)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get channel key")
	}
	sender, err := s.lookup(ctx, req.Sender, nil)
	if err != nil {
		return nil, err
	}
	senderKey, err := s.edx25519Key(sender)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get sender key")
	}

	// TODO: whitelist only keys/users we have saved?

	rids := make([]keys.ID, 0, len(req.Recipients))
	for _, invite := range req.Recipients {
		rid, err := s.lookup(ctx, invite, &lookupOpts{Verify: true})
		if err != nil {
			return nil, err
		}
		rids = append(rids, rid)
	}

	info, err := s.channelInfo(ctx, channel)
	if err != nil {
		return nil, err
	}

	msg, err := s.client.InviteToChannel(ctx, channelKey, info, senderKey, rids...)
	if err != nil {
		return nil, err
	}
	message, err := s.messageToRPC(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &ChannelInvitesCreateResponse{Message: message}, nil
}

func (s *service) ChannelUninvite(ctx context.Context, req *ChannelUninviteRequest) (*ChannelUninviteResponse, error) {
	channel, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid channel")
	}
	channelKey, err := s.edx25519Key(channel)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get channel key")
	}
	sender, err := s.lookup(ctx, req.Sender, nil)
	if err != nil {
		return nil, err
	}
	senderKey, err := s.edx25519Key(sender)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get sender key")
	}

	recipient, err := s.lookup(ctx, req.Recipient, nil)
	if err != nil {
		return nil, err
	}

	msg, err := s.client.ChannelUninvite(ctx, channelKey, senderKey, recipient)
	if err != nil {
		return nil, err
	}
	message, err := s.messageToRPC(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &ChannelUninviteResponse{Message: message}, nil
}

func (s *service) ChannelUsers(ctx context.Context, req *ChannelUsersRequest) (*ChannelUsersResponse, error) {
	user, err := s.lookup(ctx, req.User, nil)
	if err != nil {
		return nil, err
	}
	userKey, err := s.edx25519Key(user)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user key")
	}

	channel, err := s.lookup(ctx, req.Channel, nil)
	if err != nil {
		return nil, err
	}
	channelKey, err := s.edx25519Key(channel)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get channel key")
	}

	users, err := s.client.ChannelUsers(ctx, channelKey, userKey)
	if err != nil {
		return nil, err
	}

	out, err := s.channelUsersToRPC(ctx, users)
	if err != nil {
		return nil, err
	}

	return &ChannelUsersResponse{
		Users: out,
	}, nil
}

func (s *service) channelUsersToRPC(ctx context.Context, channelUsers []*api.ChannelUser) ([]*ChannelUser, error) {
	out := make([]*ChannelUser, 0, len(channelUsers))
	for _, channelUser := range channelUsers {
		key, err := s.resolveKey(ctx, channelUser.User)
		if err != nil {
			return nil, err
		}
		out = append(out, &ChannelUser{
			Key: key,
		})
	}
	return out, nil
}

func (s *service) ChannelInvites(ctx context.Context, req *ChannelInvitesRequest) (*ChannelInvitesResponse, error) {
	user, err := s.lookup(ctx, req.User, nil)
	if err != nil {
		return nil, err
	}
	userKey, err := s.edx25519Key(user)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user key")
	}

	channel, err := s.lookup(ctx, req.Channel, nil)
	if err != nil {
		return nil, err
	}
	channelKey, err := s.edx25519Key(channel)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get channel key")
	}

	invites, err := s.client.ChannelInvites(ctx, channelKey, userKey)
	if err != nil {
		return nil, err
	}

	out, err := s.invitesToRPC(ctx, invites)
	if err != nil {
		return nil, err
	}

	return &ChannelInvitesResponse{
		Invites: out,
	}, nil
}

func (s *service) ChannelUserInvites(ctx context.Context, req *ChannelUserInvitesRequest) (*ChannelUserInvitesResponse, error) {
	user, err := s.lookup(ctx, req.User, nil)
	if err != nil {
		return nil, err
	}
	userKey, err := s.edx25519Key(user)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user key")
	}

	invites, err := s.client.UserChannelInvites(ctx, userKey)
	if err != nil {
		return nil, err
	}

	out, err := s.invitesToRPC(ctx, invites)
	if err != nil {
		return nil, err
	}

	return &ChannelUserInvitesResponse{
		Invites: out,
	}, nil
}

func (s *service) invitesToRPC(ctx context.Context, invites []*api.ChannelInvite) ([]*ChannelInvite, error) {
	out := make([]*ChannelInvite, 0, len(invites))
	for _, invite := range invites {
		// channelKey, pk, err := invite.DecryptKey(s.vault)
		// if err != nil {
		// 	return nil, err
		// }
		info, pk, err := invite.DecryptInfo(s.vault)
		if err != nil {
			logger.Errorf("Invalid invite %s", invite.Channel)
			continue
		}
		channel := &Channel{
			ID:   invite.Channel.String(),
			Name: info.Name,
		}
		recipient, err := s.resolveKey(ctx, invite.Recipient)
		if err != nil {
			return nil, err
		}
		sender, err := s.resolveKey(ctx, pk.ID())
		if err != nil {
			return nil, err
		}
		out = append(out, &ChannelInvite{
			Channel:   channel,
			Recipient: recipient,
			Sender:    sender,
		})
	}
	return out, nil
}

func (s *service) ChannelJoin(ctx context.Context, req *ChannelJoinRequest) (*ChannelJoinResponse, error) {
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
	channel, _, err := invite.DecryptKey(saltpack.NewKeyring(user))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt channel key")
	}
	if _, _, err := s.vault.SaveKey(kapi.NewKey(channel)); err != nil {
		return nil, err
	}

	// Join (accept invite).
	if _, err := s.client.ChannelJoin(ctx, user, channel); err != nil {
		return nil, err
	}

	return &ChannelJoinResponse{}, nil
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
	channels, err := s.client.Channels(ctx, userKey)
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
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
			ID:   channel.String(),
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

func (s *service) channelInfo(ctx context.Context, channel keys.ID) (*api.ChannelInfo, error) {
	ch, err := s.channel(ctx, channel)
	if err != nil {
		return nil, err
	}
	return &api.ChannelInfo{Name: ch.Name}, nil
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
