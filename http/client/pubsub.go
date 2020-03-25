package client

import (
	"bytes"
	"context"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"golang.org/x/net/websocket"
)

// Publish ...
func (c *Client) Publish(ctx context.Context, sender keys.ID, recipient keys.ID, b []byte) error {
	senderKey, err := c.ks.EdX25519Key(sender)
	if err != nil {
		return err
	}
	if senderKey == nil {
		return keys.NewErrNotFound(sender.String())
	}

	sp := saltpack.NewSaltpack(c.ks)
	encrypted, err := sp.Signcrypt(b, senderKey, recipient, sender)
	if err != nil {
		return err
	}

	path := keys.Path("publish", senderKey.ID(), recipient)
	vals := url.Values{}
	if _, err := c.postDocument(ctx, path, vals, senderKey, bytes.NewReader(encrypted)); err != nil {
		return err
	}
	return nil
}

// PubSubMessage is a message from Subscribe.
type PubSubMessage struct {
	KID  keys.ID
	Data []byte
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
			var b []byte
			logger.Debugf("Receive message...")
			if err := websocket.Message.Receive(conn, &b); err != nil {
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
				logger.Debugf("Notify message...")
				receiveFn(&PubSubMessage{
					KID:  pk.ID(),
					Data: decrypted,
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
