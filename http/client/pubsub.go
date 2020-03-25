package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
)

// PubSubMessage is a message from Subscribe.
type PubSubMessage struct {
	Sender    keys.ID
	Recipient keys.ID
	Data      []byte
	Type      Pub
}

type Pub string

const (
	BinaryPub   Pub = ""
	MessagePub  Pub = "message"
	WormholePub Pub = "wormhole"
)

type message struct {
	Data []byte `json:"data"`
	Type Pub    `json:"type"`
}

// Publish ...
func (c *Client) Publish(ctx context.Context, sender keys.ID, recipient keys.ID, b []byte, typ Pub) error {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return err
	}
	if senderKey == nil {
		return keys.NewErrNotFound(sender.String())
	}

	msg := &message{
		Data: b,
		Type: typ,
	}
	mb, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	sp := saltpack.NewSaltpack(c.ks)
	encrypted, err := sp.Signcrypt(mb, senderKey, recipient, sender)
	if err != nil {
		return err
	}

	path := keys.Path("publish", senderKey.ID(), recipient)
	logger.Debugf("Publish %s", path)
	vals := url.Values{}
	if _, err := c.postDocument(ctx, path, vals, senderKey, bytes.NewReader(encrypted)); err != nil {
		return err
	}
	return nil
}

// Subscribe to messages for reciever.
// Returns channel and close function.
func (c *Client) Subscribe(ctx context.Context, reciever keys.ID, receiveFn func(*PubSubMessage)) error {
	key, err := c.ks.EdX25519Key(reciever)
	if err != nil {
		return err
	}
	path := keys.Path("subscribe", reciever)
	vals := url.Values{}
	conn, err := c.websocketGet(ctx, path, vals, key)
	if err != nil {
		return err
	}

	defer conn.Close()

	ch := make(chan []byte)

	var readErr error
	go func() {
		for {
			logger.Debugf("Receive message...")
			_, b, err := conn.ReadMessage()
			if err != nil {
				readErr = err
				return
			}
			// logger.Debugf("Received message: %s", spew.Sdump(b))
			ch <- b
		}
	}()

	for {
		select {
		case b := <-ch:
			if len(b) > 0 {
				sp := saltpack.NewSaltpack(c.ks)
				decrypted, pk, err := sp.SigncryptOpen(b)
				if err != nil {
					return err
				}
				var msg message
				if err := json.Unmarshal(decrypted, &msg); err != nil {
					return err
				}

				logger.Debugf("Notify message...")
				receiveFn(&PubSubMessage{
					Sender:    pk.ID(),
					Recipient: key.ID(),
					Data:      msg.Data,
					Type:      msg.Type,
				})
			}
		case <-ctx.Done():
			err := ctx.Err()
			if err == context.Canceled {
				return nil
			}
			return err
		case <-time.After(time.Second):
			if readErr != nil {
				return readErr
			}
		}
	}

}
