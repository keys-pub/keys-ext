package service

import (
	"io"

	"github.com/keys-pub/keysd/wormhole"
)

// Wormhole (RPC) ...
func (s *service) Wormhole(srv Keys_WormholeServer) error {
	wormhole, err := wormhole.NewWormhole(s.cfg.Server(), s.ks)
	if err != nil {
		return err
	}

	ctx := srv.Context()

	init := false
	var status WormholeStatus

	wormhole.OnOpen(func() {
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
	wormhole.OnMessage(func(b []byte) {
		if err := srv.Send(&WormholeOutput{
			Data:   b,
			Status: status,
		}); err != nil {
			logger.Errorf("Failed to send wormhole message: %v", err)
		}
	})

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
			sender, err := s.parseIdentityForEdX25519Key(ctx, req.Sender)
			if err != nil {
				return err
			}

			recipient, err := s.parseIdentityForEdX25519PublicKey(ctx, req.Recipient)
			if err != nil {
				return err
			}

			init = true
			if err := wormhole.Start(ctx, sender, recipient); err != nil {
				return err
			}
		}
		// TODO: Ensure req.Sender and req.Recipient aren't changed on subsequent requests?

		if err := wormhole.Send(req.Data); err != nil {
			return err
		}
	}

	wormhole.Close()

	return nil
}
