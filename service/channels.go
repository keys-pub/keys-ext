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

type channelsState struct {
	Channels []*api.Channel `json:"channels"`
}

func (s *service) pullChannels(ctx context.Context, user keys.ID) error {
	logger.Infof("Pull channels (%s)...", user)

	// userKey, err := s.edx25519Key(user)
	// if err != nil {
	// 	return err
	// }
	channels := []*api.Channel{}
	// channels, err := s.client.Channels(ctx, userKey)
	// if err != nil {
	// 	return err
	// }
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

// func (s *service) channelInfo(ctx context.Context, channel keys.ID) (*api.ChannelInfo, error) {
// 	ch, err := s.channel(ctx, channel)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &api.ChannelInfo{Name: ch.Name}, nil
// }

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
