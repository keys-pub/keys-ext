package stun

import (
	"time"

	"github.com/pkg/errors"
	"gortc.io/stun"
)

var stunServer = "stun.l.google.com:19302"

type Client struct {
	publicAddr stun.XORMappedAddress
	conn       *UDPConn
	onStunAddr func(addr string)
	onMessage  func(message []byte)
}

func NewClient() *Client {
	return &Client{
		onStunAddr: func(addr string) {},
		onMessage:  func([]byte) {},
	}
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) OnStunAddr(f func(addr string)) {
	c.onStunAddr = f
}

func (c *Client) OnMessage(f func(b []byte)) {
	c.onMessage = f
}

// SetPeer sets peer address.
func (c *Client) SetPeer(addr string) error {
	logger.Infof("Set peer %s", addr)
	return c.conn.SetPeer(addr)
}

// Send to peer.
func (c *Client) Send(message []byte) error {
	return c.conn.Send(message)
}

func (c *Client) Listen() error {
	conn, err := ListenUDP()
	if err != nil {
		return err
	}
	c.conn = conn
	defer c.conn.Close()

	logger.Infof("Listening on %s", c.conn.LocalAddr())

	messageChan := c.conn.Listen()

	if err := c.conn.SendBindingRequest(); err != nil {
		return err
	}

	logger.Infof("Waiting for messages...")
	for {
		select {
		case <-time.After(time.Second * 10):
			return errors.Errorf("stun timed out")
		case message, ok := <-messageChan:
			if !ok {
				logger.Infof("Listen done")
				return nil
			}
			if stun.IsMessage(message) {
				m := new(stun.Message)
				m.Raw = message
				decErr := m.Decode()
				if decErr != nil {
					return errors.Wrapf(decErr, "failed to decode stun message")
				}
				var xorAddr stun.XORMappedAddress
				if getErr := xorAddr.GetFrom(m); getErr != nil {
					return errors.Wrapf(getErr, "failed to get address from stun")
				}
				logger.Infof("Got STUN message: %s", xorAddr.String())

				if c.publicAddr.String() != xorAddr.String() {
					logger.Infof("Public address: %s", xorAddr)
					c.publicAddr = xorAddr
					c.onStunAddr(c.publicAddr.String())
				}
			} else {
				logger.Infof("Got message (%d)", len(message))
				c.onMessage(message)
			}
		}
	}
}
