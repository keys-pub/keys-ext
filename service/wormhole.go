package service

import (
	"context"
	"io"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/wormhole"
	"github.com/keys-pub/keys-ext/wormhole/sctp"
	"github.com/pkg/errors"
)

// ErrWormholeTimedOut is timed out.
var ErrWormholeTimedOut = errors.New("wormhole timed out")

func (s *service) wormholeInit(ctx context.Context, req *WormholeInput, wh *wormhole.Wormhole, srv Keys_WormholeServer) error {
	if req.ID != "" || len(req.Data) != 0 {
		return errors.Errorf("first request should not include a message")
	}

	if err := srv.Send(&WormholeOutput{Status: WormholeStarting}); err != nil {
		return err
	}

	var initiator bool
	var offer *sctp.Addr
	var sender keys.ID
	var recipient keys.ID
	if req.Invite != "" {
		if req.Sender == "" || req.Recipient != "" {
			return errors.Errorf("specify invite or sender/recipient")
		}

		invite, err := wh.FindInviteCode(ctx, req.Invite)
		if err != nil {
			return err
		}
		sender = invite.Sender
		recipient = invite.Recipient
	} else {
		if req.Sender == "" {
			return errors.Errorf("no sender specified")
		}
		sid, err := s.lookup(ctx, req.Sender, &LookupOpts{Verify: true})
		if err != nil {
			return err
		}
		sender = sid

		if req.Recipient == "" {
			return errors.Errorf("no recipient specified")
		}
		rid, err := s.lookup(ctx, req.Recipient, &LookupOpts{Verify: true})
		if err != nil {
			return err
		}
		recipient = rid
	}

	found, err := wh.FindOffer(ctx, recipient, sender)
	if err != nil {
		return errors.Wrapf(err, "failed to find offer")
	}
	if found == nil {
		initiator = true
		// created, err := wh.CreateLocalOffer(ctx, sender, recipient)
		created, err := wh.CreateOffer(ctx, sender, recipient)
		if err != nil {
			return wormholeError(err)
		}
		offer = created

		// Offering
		if err := srv.Send(&WormholeOutput{Status: WormholeOffering}); err != nil {
			return errors.Wrapf(err, "failed to offer")
		}

		// TODO: Invite
	} else {
		offer = found
		// Answering
		if err := srv.Send(&WormholeOutput{Status: WormholeAnswering}); err != nil {
			return errors.Wrapf(err, "failed to answer")
		}
	}

	if initiator {
		if err := wh.Connect(ctx, sender, recipient, offer); err != nil {
			return wormholeError(err)
		}
	} else {
		if err := wh.Listen(ctx, sender, recipient, offer); err != nil {
			return wormholeError(err)
		}
	}
	return nil
}

func wormholeError(err error) error {
	if errors.Cause(err) == context.DeadlineExceeded {
		return ErrWormholeTimedOut
	}
	return err
}

func (s *service) wormholeInput(ctx context.Context, req *WormholeInput, wh *wormhole.Wormhole) error {
	// TODO: Ensure req.Sender and req.Recipient aren't set on subsequent requests?

	if req.ID == "" {
		return errors.Errorf("no message")
	}
	_, err := wh.WriteMessage(ctx, req.ID, req.Data, wormholeContentTypeFromRPC(req.Type))
	if err != nil {
		return err
	}
	return nil
}

func (s *service) wormholeReadSend(ctx context.Context, wh *wormhole.Wormhole, srv Keys_WormholeServer) error {
	msg, err := wh.ReadMessage(ctx, true)
	if err != nil {
		return err
	}

	out, err := s.wormholeMessageToRPC(ctx, msg)
	if err != nil {
		return err
	}

	if err := srv.Send(&WormholeOutput{
		Message: out,
	}); err != nil {
		return err
	}
	return nil
}

// Wormhole (RPC) ...
func (s *service) Wormhole(srv Keys_WormholeServer) error {
	// TODO: EOF's if auth token is stale? Need better error?

	wh, err := wormhole.New(s.env.Server(), s.vault)
	if err != nil {
		return err
	}
	defer wh.Close()

	init := false

	wh.OnStatus(func(st wormhole.Status) {
		rst := wormholeStatusToRPC(st)
		if rst == WormholeDefault {
			return
		}
		if err := srv.Send(&WormholeOutput{Status: rst}); err != nil {
			logger.Errorf("Failed to send wormhole open status: %v", err)
		}
	})

	reqCh := make(chan *WormholeInput)

	ctx, cancel := context.WithCancel(srv.Context())
	defer cancel()

	var recvErr error
	go func() {
		for {
			logger.Debugf("Wormhole recv...")
			req, err := srv.Recv()
			if err == io.EOF {
				close(reqCh)
				cancel()
				return
			}
			if err != nil {
				recvErr = err
				close(reqCh)
				cancel()
				return
			}
			reqCh <- req
		}
	}()

	for req := range reqCh {
		if !init {
			init = true

			if err := s.wormholeInit(ctx, req, wh, srv); err != nil {
				return err
			}

			go func() {
				for {
					if err := s.wormholeReadSend(ctx, wh, srv); err != nil {
						return
					}
				}
			}()

		} else {
			if err := s.wormholeInput(ctx, req, wh); err != nil {
				return err
			}
		}
	}
	if recvErr != nil {
		return recvErr
	}
	return nil

}

func wormholeStatusToRPC(st wormhole.Status) WormholeStatus {
	switch st {
	case wormhole.SCTPHandshake:
		return WormholeHandshake
	case wormhole.Connected:
		return WormholeConnected
	case wormhole.Closed:
		return WormholeClosed
	default:
		return WormholeDefault
	}
}

func wormholeContentTypeFromRPC(typ ContentType) wormhole.ContentType {
	switch typ {
	case UTF8Content:
		return wormhole.UTF8Content
	default:
		return wormhole.BinaryContent
	}
}

func wormholeMessageTypeToRPC(typ wormhole.MessageType) WormholeMessageType {
	switch typ {
	case wormhole.Sent:
		return WormholeMessageSent
	case wormhole.Pending:
		return WormholeMessagePending
	case wormhole.Ack:
		return WormholeMessageAck
	default:
		// TODO:
		return WormholeMessageSent
	}
}

func wormholeContentTypeToRPC(typ wormhole.ContentType) ContentType {
	switch typ {
	case wormhole.UTF8Content:
		return UTF8Content
	default:
		return BinaryContent
	}
}

func (s *service) wormholeMessageToRPC(ctx context.Context, msg *wormhole.Message) (*WormholeMessage, error) {
	sender, err := s.key(ctx, msg.Sender)
	if err != nil {
		return nil, err
	}

	out := &WormholeMessage{
		ID: msg.ID,
		Content: &Content{
			Data: msg.Content.Data,
			Type: wormholeContentTypeToRPC(msg.Content.Type),
		},
		Type:   wormholeMessageTypeToRPC(msg.Type),
		Sender: sender,
	}
	return out, nil
}
