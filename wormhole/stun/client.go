package stun

import (
	"net"

	"github.com/pkg/errors"
	"gortc.io/stun"
)

var stunServer = "stun.l.google.com:19302"

type OnPeer func(addr string)
type OnMessage func(message []byte)

type Conn interface {
	Send(message []byte) error
	Listen() <-chan []byte
	LocalAddr() net.Addr
	SendBindingRequest() error
	SetPeer(addr string) error
	Close()
}

type Client struct {
	publicAddr stun.XORMappedAddress
	conn       Conn
	quit       chan bool
	onPeer     OnPeer
	onMessage  OnMessage
}

func NewClient() *Client {
	return &Client{
		onPeer:    func(addr string) {},
		quit:      make(chan bool),
		onMessage: func([]byte) {},
	}
}

func (c *Client) Close() {
	c.quit <- true
}

func (c *Client) OnPeer(f OnPeer) {
	c.onPeer = f
}

func (c *Client) OnMessage(f OnMessage) {
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
	// keepAlive := time.NewTicker(time.Second)

	if err := c.conn.SendBindingRequest(); err != nil {
		return err
	}

	logger.Infof("Waiting for messages...")
	for {
		select {
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
					c.onPeer(c.publicAddr.String())
				}
			} else {
				logger.Infof("Got message (%d)", len(message))
				c.onMessage(message)
			}
		// case <-keepAlive.C:
		// 	// Keep alive
		// 	logger.Infof("Keep alive...")
		case <-c.quit:
			logger.Infof("Closing connection...")
			c.conn.Close()
		}
	}
}
