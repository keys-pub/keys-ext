package service

import (
	"context"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	kapi "github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/pkg/errors"
)

func (s *service) Channels(ctx context.Context, req *ChannelsRequest) (*ChannelsResponse, error) {
	if err := s.pullChannels(ctx); err != nil {
		return nil, err
	}

	channels, err := s.channels(ctx)
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
	key := keys.GenerateEdX25519Key()

	info := &api.ChannelInfo{Name: name}
	created, err := s.client.ChannelCreate(ctx, key, userKey, info)
	if err != nil {
		return nil, err
	}

	ck := kapi.NewKey(key).Created(s.clock.NowMillis()).WithLabel("channel")
	ck.Token = created.Channel.Token
	if err := s.vault.SaveKey(ck); err != nil {
		return nil, err
	}

	return &ChannelCreateResponse{
		Channel: &Channel{
			ID: ck.ID.String(),
		},
	}, nil
}

func (s *service) pullChannels(ctx context.Context) error {
	logger.Infof("Pull channels...")

	cks, err := s.channelKeys()
	if err != nil {
		return err
	}
	tokens := []*api.ChannelToken{}
	for _, ck := range cks {
		token := &api.ChannelToken{
			ID:    ck.ID,
			Token: ck.Token,
		}
		tokens = append(tokens, token)
	}
	remoteStatus, err := s.client.ChannelsStatus(ctx, tokens...)
	if err != nil {
		return err
	}
	logger.Debugf("Channels status (remote): %s", spew.Sdump(remoteStatus))
	remoteStatusBy := map[keys.ID]*api.ChannelStatus{}
	for _, rs := range remoteStatus {
		remoteStatusBy[rs.ID] = rs
	}

	// TODO: Pull channels in a single (bulk) call
	for _, ck := range cks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			remoteStatus := remoteStatusBy[ck.ID]
			status, err := s.channelStatus(ctx, ck.ID)
			if err != nil {
				return err
			}
			if remoteStatus != nil && status.Index < remoteStatus.Index {
				if err := s.pullMessages(ctx, ck); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type channelStatus struct {
	ID              keys.ID `json:"id,omitempty" msgpack:"id,omitempty"`
	Name            string  `json:"name,omitempty" msgpack:"name,omitempty"`
	Description     string  `json:"desc,omitempty" msgpack:"desc,omitempty"`
	Snippet         string  `json:"snippet,omitempty" msgpack:"snippet,omitempty"`
	Index           int64   `json:"index,omitempty" msgpack:"index,omitempty"`
	Timestamp       int64   `json:"ts,omitempty" msgpack:"ts,omitempty"`
	RemoteTimestamp int64   `json:"rts,omitempty" msgpack:"rts,omitempty"`
}

func (s *service) channelStatus(ctx context.Context, cid keys.ID) (*channelStatus, error) {
	// channelStatus is set during pullMessages
	var cs channelStatus
	ok, err := s.db.Load(ctx, dstore.Path("channels", cid), &cs)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &channelStatus{
			ID:   cid,
			Name: cid.String(),
		}, nil
	}
	return &cs, nil
}

func (s *service) channels(ctx context.Context) ([]*Channel, error) {
	docs, err := s.db.Documents(ctx, dstore.Path("channels"))
	if err != nil {
		return nil, err
	}
	channels := []*Channel{}
	for _, doc := range docs {
		var channelStatus channelStatus
		if err := doc.To(&channelStatus); err != nil {
			return nil, err
		}
		channels = append(channels, channelToRPC(channelStatus))
	}
	return channels, nil
}

// func (s *service) channel(ctx context.Context, id keys.ID) (*Channel, error) {
// 	var channelStatus channelStatus
// 	ok, err := s.db.Load(ctx, dstore.Path("channels", id), &channelStatus)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !ok {
// 		return nil, nil
// 	}
// 	return channelToRPC(channelStatus), nil
// }

func channelToRPC(channelStatus channelStatus) *Channel {
	return &Channel{
		ID:        channelStatus.ID.String(),
		Name:      channelStatus.Name,
		Snippet:   channelStatus.Snippet,
		UpdatedAt: channelStatus.RemoteTimestamp,
		Index:     channelStatus.Index,
	}
}

func (s *service) channelKeys() ([]*kapi.Key, error) {
	ks, err := s.vault.Keys()
	if err != nil {
		return nil, err
	}
	out := []*kapi.Key{}
	for _, key := range ks {
		if !key.HasLabel("channel") {
			continue
		}
		out = append(out, key)
	}
	logger.Debugf("Found %d channels", len(out))
	return out, nil
}
