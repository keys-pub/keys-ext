package service

import (
	"context"

	"github.com/keys-pub/keys"
)

func (s *service) Follow(ctx context.Context, req *FollowRequest) (*FollowResponse, error) {
	recipient, err := s.lookup(ctx, req.Recipient, nil)
	if err != nil {
		return nil, err
	}
	senderKey, err := s.lookupKey(ctx, req.Sender, nil)
	if err != nil {
		return nil, err
	}
	if err := s.client.Follow(ctx, senderKey.AsEdX25519(), recipient); err != nil {
		return nil, err
	}

	return &FollowResponse{
		Follow: &Follow{
			Recipient: recipient.String(),
			Sender:    senderKey.ID.String(),
		},
	}, nil
}

func (s *service) Follows(ctx context.Context, req *FollowsRequest) (*FollowsResponse, error) {
	recipient, err := keys.ParseID(req.Recipient)
	if err != nil {
		return nil, err
	}
	recipientKey, err := s.vaultKey(recipient)
	if err != nil {
		return nil, err
	}

	follows, err := s.client.Follows(ctx, recipientKey.AsEdX25519())
	if err != nil {
		return nil, err
	}

	out := []*Follow{}
	for _, follow := range follows {
		out = append(out, &Follow{Sender: follow.Sender.String(), Recipient: follow.Recipient.String()})
	}

	return &FollowsResponse{
		Follows: out,
	}, nil
}
