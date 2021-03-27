package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/vault/keyring"
	kapi "github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore"
)

type dmStatus struct {
	Index int64 `json:"index,omitempty" msgpack:"index,omitempty"`
}

type directMessages struct {
	Tokens []string
}

func (s *service) pullDirectMessages(ctx context.Context, userKey *kapi.Key) (*directMessages, error) {
	dms := &directMessages{Tokens: []string{}}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			dmns, err := s.pullDirectMessagesNext(ctx, userKey)
			if err != nil {
				return nil, err
			}
			dms.Tokens = append(dms.Tokens, dmns.Tokens...)
			if !dmns.Truncated {
				return dms, nil
			}
		}
	}
}

type directMessagesNext struct {
	Truncated bool
	Tokens    []string
}

func (s *service) pullDirectMessagesNext(ctx context.Context, userKey *kapi.Key) (*directMessagesNext, error) {
	logger.Infof("Pull dms (%s)...", userKey.ID)

	status, err := s.dmStatus(ctx, userKey.ID)
	if err != nil {
		return nil, err
	}

	logger.Infof("Pull dms from %d", status.Index)
	opts := &client.MessagesOpts{
		Index: status.Index,
	}
	key := userKey.AsEdX25519()
	directs, err := s.client.DirectMessages(ctx, key, opts)
	if err != nil {
		return nil, err
	}
	kr := keyring.New(s.vault)
	tokens := []string{}
	for _, event := range directs.Events {
		msg, err := api.DecryptMessageFromEvent(event, key)
		if err != nil {
			return nil, err
		}
		path := dstore.Path("dms", userKey.ID, pad(event.Index))
		if err := s.db.Set(ctx, path, dstore.From(event)); err != nil {
			return nil, err
		}
		// If channel invite, add key
		for _, invite := range msg.ChannelInvites {
			key := invite.Key.WithLabels("channel")
			key.Token = invite.Token
			if err := kr.Save(key); err != nil {
				return nil, err
			}
			tokens = append(tokens, invite.Token)
		}
	}
	status.Index = directs.Index

	// Save status
	if err := s.db.Set(ctx, dstore.Path("dms", userKey.ID), dstore.From(status), dstore.MergeAll()); err != nil {
		return nil, err
	}

	return &directMessagesNext{
		Truncated: directs.Truncated,
		Tokens:    tokens,
	}, nil
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
