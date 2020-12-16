package service

import (
	"context"
	"strings"

	"github.com/pkg/errors"
)

func (s *service) userCommand(ctx context.Context, cmd string, user string) error {
	fields := strings.Fields(cmd)
	if len(cmd) == 0 {
		return errors.Errorf("no command")
	}

	cmd0, args := fields[0], fields[1:]

	switch cmd0 {
	case "/create":
		_, err := s.ChannelCreate(ctx, &ChannelCreateRequest{User: user, Name: args[0]})
		if err != nil {
			return err
		}
		return nil

	case "/join":
		_, err := s.ChannelJoin(ctx, &ChannelJoinRequest{User: user, Channel: args[0]})
		if err != nil {
			return err
		}
		return nil
	}
	return errors.Errorf("unrecognized command")
}

func (s *service) channelCommand(ctx context.Context, cmd string, user string, channel string) (*Message, error) {
	fields := strings.Fields(cmd)
	if len(cmd) == 0 {
		return nil, errors.Errorf("no command")
	}

	cmd0, args := fields[0], fields[1:]

	switch cmd0 {
	case "/invite":
		if len(args) == 0 {
			return nil, errors.Errorf("no recipients")
		}
		resp, err := s.ChannelInvitesCreate(ctx, &ChannelInvitesCreateRequest{Channel: channel, Sender: user, Recipients: args})
		if err != nil {
			return nil, err
		}
		return resp.Message, nil
	case "/uninvite":
		if len(args) < 0 {
			return nil, errors.Errorf("no recipients")
		}
		resp, err := s.ChannelUninvite(ctx, &ChannelUninviteRequest{Channel: channel, Sender: user, Recipient: args[0]})
		if err != nil {
			return nil, err
		}
		return resp.Message, nil
	case "/leave":
		resp, err := s.ChannelLeave(ctx, &ChannelLeaveRequest{Channel: channel, User: user})
		if err != nil {
			return nil, err
		}
		return resp.Message, nil
	}

	return nil, errors.Errorf("unrecognized command")
}
