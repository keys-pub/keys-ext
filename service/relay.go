package service

import (
	wsapi "github.com/keys-pub/keys-ext/ws/api"
	wsclient "github.com/keys-pub/keys-ext/ws/client"
	"github.com/pkg/errors"
)

// Relay (RPC) ...
func (s *service) Relay(req *RelayRequest, srv Keys_RelayServer) error {
	ctx := srv.Context()

	sks, err := s.lookupEdX25519Keys(ctx, req.Keys)
	if err != nil {
		return err
	}
	if len(sks) == 0 {
		return errors.Errorf("no keys specified for relay")
	}

	relay, err := wsclient.New("wss://relay.keys.pub/ws")
	if err != nil {
		return err
	}

	for _, key := range sks {
		relay.Authorize(key)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		// TODO: Context to read
		events, err := relay.ReadEvents()
		if err != nil {
			return err
		}
		for _, event := range events {
			switch event.Type {
			case wsapi.HelloEventType:
				logger.Infof("Relay hello %s", event.User)
				if err := s.pullChannels(ctx, event.User); err != nil {
					return err
				}
			case wsapi.ChannelMessageEventType:
				logger.Infof("Relay message %s", event.Channel)
				if err := s.pullMessages(ctx, event.Channel, event.User); err != nil {
					return err

				}
			case wsapi.ChannelCreatedEventType:
				logger.Infof("Relay channel created %s", event.Channel)
				// TODO: This pulls all channels, not just the new one.
				if err := s.pullChannels(ctx, event.User); err != nil {
					return err
				}
			}
		}
		for _, event := range events {
			var out *RelayOutput
			switch event.Type {
			case wsapi.HelloEventType:
				out = &RelayOutput{
					Type: RelayHello,
					User: event.User.String(),
				}
			case wsapi.ChannelCreatedEventType:
				out = &RelayOutput{
					Type:    RelayChannelCreated,
					User:    event.User.String(),
					Channel: event.Channel.String(),
				}
			case wsapi.ChannelMessageEventType:
				out = &RelayOutput{
					Type:    RelayChannelMessage,
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
