package service

import (
	"io"

	"github.com/keys-pub/keysd/wormhole"
	"github.com/pkg/errors"
)

// Wormhole (RPC) ...
func (s *service) Wormhole(srv Keys_WormholeServer) error {
	// TODO: EOF's if auth token is stale, need better error

	wormhole, err := wormhole.NewWormhole(s.cfg.Server(), s.ks)
	if err != nil {
		return err
	}
	defer wormhole.Close()

	ctx := srv.Context()

	init := false
	var status WormholeStatus

	wormhole.OnConnect(func() {
		status = WormholeStatusOpen
		if err := srv.Send(&WormholeOutput{
			Status: status,
		}); err != nil {
			logger.Errorf("Failed to send wormhole open status: %v", err)
		}
	})
	wormhole.OnClose(func() {
		status = WormholeStatusClosed
		if err := srv.Send(&WormholeOutput{
			Status: status,
		}); err != nil {
			logger.Errorf("Failed to send wormhole closed status: %v", err)
		}
	})

	var readErr error
	var relayErr error
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, recvErr := srv.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return recvErr
		}

		if !init {
			if req.Sender == "" {
				return errors.Errorf("no sender specified")
			}
			sender, err := s.parseIdentityForEdX25519Key(ctx, req.Sender)
			if err != nil {
				return err
			}

			if req.Recipient == "" {
				return errors.Errorf("no recipient specified")
			}
			recipient, err := s.parseIdentityForEdX25519PublicKey(ctx, req.Recipient)
			if err != nil {
				return err
			}

			init = true
			if err := wormhole.Start(ctx, sender, recipient); err != nil {
				return err
			}

			// Read and send output to client
			go func() {
				for {
					b, err := wormhole.Read(ctx)
					if err != nil {
						readErr = err
						break
					}
					if err := srv.Send(&WormholeOutput{
						Data:   b,
						Status: status,
					}); err != nil {
						relayErr = err
						break
					}
				}
			}()
		}
		// TODO: Ensure req.Sender and req.Recipient aren't changed on subsequent requests?

		if readErr != nil {
			return readErr
		}
		if relayErr != nil {
			return relayErr
		}

		if err := wormhole.Send(ctx, req.Data); err != nil {
			return err
		}
	}

	return nil
}
