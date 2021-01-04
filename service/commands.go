package service

import (
	"context"
	"strings"

	"github.com/pkg/errors"
)

func (s *service) command(ctx context.Context, cmd string, user string, channel string) (*Message, error) {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return nil, errors.Errorf("no command")
	}

	cmd0, args := fields[0], fields[1:]
	logger.Debugf("Channel command: %s %v", cmd0, args)

	switch cmd0 {
	case "/invite":
		resp, err := s.ChannelInvite(ctx, &ChannelInviteRequest{Channel: channel, Sender: user, Recipients: args})
		if err != nil {
			return nil, err
		}
		return resp.Message, nil
	case "/leave":
		return nil, errors.Errorf("not implemented")
	case "/create":
		_, err := s.ChannelCreate(ctx, &ChannelCreateRequest{Name: args[0], User: user})
		if err != nil {
			return nil, err
		}
		return nil, nil
	case "/follow":
		_, err := s.Follow(ctx, &FollowRequest{Recipient: args[0], Sender: user})
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	return nil, errors.Errorf("unrecognized command")
}
