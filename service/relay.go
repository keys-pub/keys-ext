package service

import (
	"time"

	"github.com/keys-pub/keys-ext/ws/api"
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
					// TODO: Channel user
					// if err := s.pullMessages(ctx, event.Channel, keys.ID("")); err != nil {
					// 	return err
					// }
					return errors.Errorf("not implemented")
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
