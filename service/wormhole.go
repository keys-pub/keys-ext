package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/wormhole"
	"github.com/keys-pub/keysd/wormhole/sctp"
	"github.com/pkg/errors"
)

// Wormhole (RPC) ...
func (s *service) Wormhole(srv Keys_WormholeServer) error {
	// TODO: EOF's if auth token is stale, need better error

	wh, err := wormhole.NewWormhole(s.cfg.Server(), s.ks)
	if err != nil {
		return err
	}
	defer wh.Close()

	ctx, cancel := context.WithCancel(srv.Context())
	defer cancel()

	init := false
	var status WormholeStatus

	wh.OnConnect(func() {
		status = WormholeStatusOpen
		if err := srv.Send(&WormholeOutput{
			Status: status,
		}); err != nil {
			logger.Errorf("Failed to send wormhole open status: %v", err)
		}
	})
	wh.OnClose(func() {
		status = WormholeStatusClosed
		if err := srv.Send(&WormholeOutput{
			Status: status,
		}); err != nil {
			logger.Errorf("Failed to send wormhole closed status: %v", err)
		}
	})

	var readErr error
	var startErr error
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
			if req.ID != "" || len(req.Data) != 0 {
				return errors.Errorf("first request should only be sender/recipient")
			}

			init = true

			var initiator bool
			var offer *sctp.Addr
			var sender keys.ID
			var recipient keys.ID
			if req.Invite == "" {
				if req.Sender == "" {
					return errors.Errorf("no sender specified")
				}
				sid, err := s.parseIdentity(ctx, req.Sender)
				if err != nil {
					return err
				}
				sender = sid

				if req.Recipient == "" {
					return errors.Errorf("no recipient specified")
				}
				rid, err := s.parseIdentity(ctx, req.Recipient)
				if err != nil {
					return err
				}
				recipient = rid

				found, err := wh.FindOffer(ctx, sender, recipient)
				if err != nil {
					return err
				}
				if found == nil {
					initiator = true
					created, invite, err := wh.CreateOffer(ctx, sender, recipient)
					if err != nil {
						return err
					}
					fmt.Printf("Invite code: %s\n", invite)
					offer = created
				} else {
					offer = found
				}
			}

			go func() {
				if req.Invite != "" {
					if err := wh.ListenByInvite(ctx, req.Invite); err != nil {
						startErr = err
						return
					}
				} else if initiator {
					if err := wh.Connect(ctx, sender, recipient, offer); err != nil {
						startErr = err
						return
					}
				} else {
					if err := wh.Listen(ctx, sender, recipient, offer); err != nil {
						startErr = err
						return
					}
				}

				// Read and send output to client
				go func() {
					for {
						msg, err := wh.ReadMessage(ctx, true)
						if err != nil {
							readErr = err
							return
						}

						out, err := s.messageToRPC(ctx, msg)
						if err != nil {
							readErr = err
							return
						}

						if err := srv.Send(&WormholeOutput{
							Message: out,
							Status:  status,
						}); err != nil {
							readErr = err
							return
						}
					}
				}()
			}()
			continue
		}
		// TODO: Ensure req.Sender and req.Recipient aren't changed on subsequent requests?

		if readErr != nil {
			return readErr
		}
		if startErr != nil {
			return startErr
		}
		if req.ID != "" {
			_, err := wh.WriteMessage(ctx, req.ID, req.Data, contentTypeFromRPC(req.Type))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func contentTypeFromRPC(typ ContentType) wormhole.ContentType {
	switch typ {
	case UTF8Content:
		return wormhole.UTF8Content
	default:
		return wormhole.BinaryContent
	}
}

func messageTypeToRPC(typ wormhole.MessageType) MessageType {
	switch typ {
	case wormhole.Sent:
		return MessageSent
	case wormhole.Pending:
		return MessagePending
	case wormhole.Ack:
		return MessageAck
	default:
		// TODO:
		return MessageSent
	}
}

func contentTypeToRPC(typ wormhole.ContentType) ContentType {
	switch typ {
	case wormhole.UTF8Content:
		return UTF8Content
	default:
		return BinaryContent
	}
}

func (s *service) messageToRPC(ctx context.Context, msg *wormhole.Message) (*Message, error) {
	out := &Message{
		ID: msg.ID,
		Content: &Content{
			Data: msg.Content.Data,
			Type: contentTypeToRPC(msg.Content.Type),
		},
		Type: messageTypeToRPC(msg.Type),
	}

	if err := s.fillMessage(ctx, out, time.Time{}, msg.Sender); err != nil {
		return nil, err
	}
	return out, nil
}
