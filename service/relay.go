package service

import (
	"time"

	"github.com/keys-pub/keys-ext/ws/api"
	wsclient "github.com/keys-pub/keys-ext/ws/client"
)

// Relay (RPC) ...
func (s *service) Relay(req *RelayRequest, srv Keys_RelayServer) error {
	ctx := srv.Context()

	relay, err := wsclient.New("wss://relay.keys.pub/ws")
	if err != nil {
		return err
	}

	cks, err := s.channelKeys()
	if err != nil {
		return err
	}

	tokens := []string{}
	for _, ck := range cks {
		if ck.Token != "" {
			tokens = append(tokens, ck.Token)
		}
	}
	relay.Authorize(tokens)

	if err := relay.Connect(); err != nil {
		return err
	}
	defer relay.Close()

	chEvents := make(chan []*api.Event)
	chErr := make(chan error)

	go func() {
		for {
			logger.Infof("Read relay events...")
			events, err := relay.ReadEvents()
			if err != nil {
				chErr <- err
				return
			}
			chEvents <- events
		}
	}()

	ticker := time.NewTicker(50 * time.Second)

	for {
		select {
		case <-ticker.C:
			if err := relay.SendPing(); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		case err := <-chErr:
			return err
		case events := <-chEvents:
			logger.Infof("Got relay events...")
			for _, event := range events {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					logger.Infof("Relay event %s", event.Channel)
					ck, err := s.vaultKey(event.Channel)
					if err != nil {
						return err
					}
					if ck == nil {
						logger.Infof("Channel key not found: %s", event.Channel)
						continue
					}
					if err := s.pullMessages(ctx, ck); err != nil {
						return err
					}
				}
			}
			for _, event := range events {
				out := &RelayOutput{
					Channel: event.Channel.String(),
				}
				if err := srv.Send(out); err != nil {
					return err
				}
			}
		}
	}
}
