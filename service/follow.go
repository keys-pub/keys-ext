package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

func (s *service) Follow(ctx context.Context, req *FollowRequest) (*FollowResponse, error) {
	recipient, err := keys.ParseID(req.Recipient)
	if err != nil {
		return nil, err
	}
	sender, err := keys.ParseID(req.Sender)
	if err != nil {
		return nil, err
	}
	senderKey, err := s.vaultKey(sender)
	if err != nil {
		return nil, err
	}
	token := senderKey.Token
	if token == "" {
		return nil, errors.Errorf("no token for sender")
	}

	if err := s.client.Follow(ctx, senderKey.AsEdX25519(), recipient, token); err != nil {
		return nil, err
	}

	return &FollowResponse{
		Follow: &Follow{
			Recipient: recipient.String(),
			Sender:    sender.String(),
		},
	}, nil
}
