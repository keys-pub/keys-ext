package service

import (
	wsapi "github.com/keys-pub/keys-ext/ws/api"
	wsclient "github.com/keys-pub/keys-ext/ws/client"
)

// NotifyStream (RPC) ...
func (s *service) NotifyStream(req *NotifyStreamRequest, srv Keys_NotifyStreamServer) error {
	ctx := srv.Context()

	cl, err := wsclient.New("wss://relay.keys.pub/ws")
	if err != nil {
		return err
	}
	sks, err := s.vault.EdX25519Keys()
	if err != nil {
		return err
	}
	for _, sk := range sks {
		user, err := s.user(ctx, sk.ID())
		if err != nil {
			return err
		}
		if user != nil {
			cl.Authorize(sk)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		events, err := cl.ReadEvents()
		if err != nil {
			return err
		}
		for _, event := range events {
			switch event.Type {
			case wsapi.ChannelMessageEventType:
				if err := s.pullMessages(ctx, event.Channel, event.User); err != nil {
					return err
				}
			case wsapi.ChannelCreatedEventType:
				// TODO: This pulls all channels, not just the new one.
				if err := s.pullChannels(ctx, event.User); err != nil {
					return err
				}
			}
		}

		for _, event := range events {
			var out *NotifyStreamOutput
			switch event.Type {
			case wsapi.ChannelCreatedEventType:
				out = &NotifyStreamOutput{
					Type:    ChannelCreatedNotification,
					User:    event.User.String(),
					Channel: event.Channel.String(),
				}
			case wsapi.ChannelMessageEventType:
				out = &NotifyStreamOutput{
					Type:    ChannelMessageNotification,
					Channel: event.Channel.String(),
					User:    event.User.String(),
					Index:   event.Index,
				}
			default:
				continue
			}

			if err := srv.Send(out); err != nil {
				return err
			}
		}
	}
}
