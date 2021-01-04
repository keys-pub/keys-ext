package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	kapi "github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore"
)

type dmStatus struct {
	Index int64 `json:"index,omitempty" msgpack:"index,omitempty"`
}

func (s *service) pullDirectMessages(ctx context.Context, userKey *kapi.Key) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			truncated, err := s.pullDirectMessagesNext(ctx, userKey)
			if err != nil {
				return err
			}
			if !truncated {
				return nil
			}
		}
	}
}

func (s *service) pullDirectMessagesNext(ctx context.Context, userKey *kapi.Key) (bool, error) {
	logger.Infof("Pull dms (%s)...", userKey.ID)

	status, err := s.dmStatus(ctx, userKey.ID)
	if err != nil {
		return false, err
	}

	logger.Infof("Pull dms from %d", status.Index)
	opts := &client.MessagesOpts{
		Index: status.Index,
	}
	key := userKey.AsEdX25519()
	directs, err := s.client.DirectMessages(ctx, key, opts)
	if err != nil {
		return false, err
	}
	for _, event := range directs.Events {
		msg, err := api.DecryptMessageFromEvent(event, key)
		if err != nil {
			return false, err
		}
		path := dstore.Path("dms", userKey.ID, pad(event.Index))
		if err := s.db.Set(ctx, path, dstore.From(event)); err != nil {
			return false, err
		}
		// If channel invite, add key
		for _, invite := range msg.ChannelInvites {
			key := invite.Key.WithLabel("channel")
			key.Token = invite.Token
			if err := s.vault.SaveKey(key); err != nil {
				return false, err
			}
		}
	}
	status.Index = directs.Index

	// Save status
	if err := s.db.Set(ctx, dstore.Path("dms", userKey.ID), dstore.From(status), dstore.MergeAll()); err != nil {
		return false, err
	}

	return directs.Truncated, nil
}

func (s *service) dmStatus(ctx context.Context, user keys.ID) (*dmStatus, error) {
	var ds dmStatus
	ok, err := s.db.Load(ctx, dstore.Path("dms", user), &ds)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &dmStatus{
			Index: 0,
		}, nil
	}
	return &ds, nil
}
