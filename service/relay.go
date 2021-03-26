package service

import (
	"context"
	"sync"
	"time"

	"github.com/keys-pub/keys-ext/ws/api"
	wsclient "github.com/keys-pub/keys-ext/ws/client"
	"github.com/pkg/errors"
)

type relayClient struct {
	ui RPC_RelayServer
	ws *wsclient.Client
}

type relay struct {
	sync.Mutex
	client *relayClient
}

func newRelay() *relay {
	return &relay{}
}

func (r *relay) Register(client *relayClient) {
	r.Lock()
	defer r.Unlock()
	r.client = client
}

func (r *relay) Unregister(client *relayClient) {
	r.Lock()
	defer r.Unlock()
	if client == r.client {
		r.client = nil
	}
}

func (r *relay) Send(out *RelayOutput) {
	r.Lock()
	defer r.Unlock()
	if r.client != nil {
		if err := r.client.ui.Send(out); err != nil {
			logger.Errorf("Failed to relay event: %v", err)
		}
	}
}

func (r *relay) Authorize(tokens []string) {
	r.Lock()
	defer r.Unlock()
	if r.client != nil {
		if err := r.client.ws.Authorize(tokens); err != nil {
			logger.Errorf("Failed to relay auth: %v", err)
		}
	}
}

// Relay (RPC) ...
func (s *service) Relay(req *RelayRequest, srv RPC_RelayServer) error {
	ctx := srv.Context()

	relay, err := wsclient.New("wss://relay.keys.pub/ws")
	if err != nil {
		return err
	}

	if err := relay.Connect(); err != nil {
		return err
	}
	defer relay.Close()

	client := &relayClient{ui: srv, ws: relay}
	s.relay.Register(client)
	defer s.relay.Unregister(client)

	tokens, err := s.relayTokens(ctx, req.User)
	if err != nil {
		return err
	}

	if err := relay.Authorize(tokens); err != nil {
		return err
	}

	// Send empty message to ui after connect and auth
	if err := srv.Send(&RelayOutput{}); err != nil {
		return err
	}

	chEvents := make(chan []*api.Event)

	go func() {
		for {
			logger.Infof("Read relay events...")
			events, err := relay.ReadEvents()
			if err != nil {
				logger.Errorf("Error reading events: %v", err)
				relay.Close()
				return
			}
			chEvents <- events
		}
	}()

	ticker := time.NewTicker(50 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			logger.Debugf("Send ping...")
			if err := relay.Ping(); err != nil {
				return err
			}
		case events := <-chEvents:
			logger.Infof("Got relay events...")
			for _, event := range events {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					logger.Infof("Relay event %v", event)
					if event.Channel != "" {
						ck, err := s.vaultKey(event.Channel)
						if err != nil {
							return err
						}
						if ck == nil {
							logger.Infof("Channel key not found: %s", event.Channel)
							continue
						}
						// TODO: Pull messages
					}
					if event.User != "" {
						uk, err := s.vaultKey(event.User)
						if err != nil {
							return err
						}
						if uk == nil {
							logger.Infof("User key not found: %s", event.User)
							continue
						}
						// TODO: Pull direct messages

						// if err := relay.Authorize(dms.Tokens); err != nil {
						// 	return err
						// }
					}
				}
			}
			for _, event := range events {
				out := &RelayOutput{
					Channel: event.Channel.String(),
					User:    event.User.String(),
				}
				if err := srv.Send(out); err != nil {
					return err
				}
			}
		}
	}
}

func (s *service) relayTokens(ctx context.Context, user string) ([]string, error) {
	return nil, errors.Errorf("not implemented")
}
