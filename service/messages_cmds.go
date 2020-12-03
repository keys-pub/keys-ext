package service

import (
	"context"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/pkg/errors"
)

func (s *service) inviteToChannel(ctx context.Context, keys []string, channel *keys.EdX25519Key, user *keys.EdX25519Key) (*api.Message, error) {
	kids, err := s.lookupAll(ctx, keys, nil)
	if err != nil {
		return nil, err
	}

	// TODO: whitelist only keys/users we have saved.

	msg := api.NewMessage()
	msg.ChannelInviteNn = &api.ChannelInviteNn{Recipients: kids, Sender: user.ID()}
	msg.Timestamp = s.clock.NowMillis()

	if err := s.client.InviteToChannel(ctx, channel, user, kids...); err != nil {
		return nil, err
	}

	if err := s.client.MessageSend(ctx, msg, user, channel); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *service) messageCommand(ctx context.Context, cmd string, channel *keys.EdX25519Key, user *keys.EdX25519Key) (*Message, error) {
	fields := strings.Fields(cmd)
	if len(cmd) == 0 {
		return nil, errors.Errorf("no command")
	}

	switch fields[0] {
	case "/invite":
		msg, err := s.inviteToChannel(ctx, fields[1:], channel, user)
		if err != nil {
			return nil, err
		}
		out, err := s.messageToRPC(ctx, msg)
		if err != nil {
			return nil, err
		}
		return out, nil
	}

	return nil, errors.Errorf("unrecognized command")

}
