package wormhole

import (
	"net"

	"github.com/pkg/errors"
	"gortc.io/stun"
)

var stunServer = "stun.l.google.com:19302"
var udp = "udp4"

type PublicAddressLn func(addr string)
type MessageLn func(message []byte)

type Client struct {
	publicAddr      stun.XORMappedAddress
	peerAddr        *net.UDPAddr
	conn            *net.UDPConn
	quit            chan bool
	publicAddressLn PublicAddressLn
	messageLn       MessageLn
}

func NewClient() *Client {
	return &Client{
		publicAddressLn: func(addr string) {},
		quit:            make(chan bool),
		messageLn:       func([]byte) {},
	}
}

func (c *Client) Close() {
	c.quit <- true
}

func (c *Client) SetPublicAddressLn(publicAddressLn PublicAddressLn) {
	c.publicAddressLn = publicAddressLn
}

func (c *Client) SetMessageLn(messageLn MessageLn) {
	c.messageLn = messageLn
}

// SetPeer sets peer address.
func (c *Client) SetPeer(addr string) error {
	logger.Infof("Set peer %s", addr)
	a, err := net.ResolveUDPAddr(udp, addr)
	if err != nil {
		return err
	}
	c.peerAddr = a
	return nil
}

// Send to peer.
func (c *Client) Send(message []byte) error {
	return send(message, c.conn, c.peerAddr)
}

func (c *Client) Listen() error {
	conn, err := net.ListenUDP(udp, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to net.Listen")
	}
	c.conn = conn
	defer c.conn.Close()

	logger.Infof("Listening on %s", c.conn.LocalAddr())

	messageChan := listen(c.conn)
	// keepAlive := time.NewTicker(time.Second)

	logger.Infof("Send binding request...")
	srvAddr, err := net.ResolveUDPAddr(udp, stunServer)
	if err != nil {
		return errors.Wrapf(err, "failed to resolve addr")
	}
	if err := sendBindingRequest(c.conn, srvAddr); err != nil {
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
					c.publicAddressLn(c.publicAddr.String())
				}
			} else {
				logger.Infof("Got message (%d)", len(message))
				c.messageLn(message)
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

func listen(conn *net.UDPConn) <-chan []byte {
	messages := make(chan []byte)
	go func() {
		for {
			buf := make([]byte, 1024)

			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				// logger.Infof("Error reading: %v", err)
				close(messages)
				return
			}
			buf = buf[:n]

			messages <- buf
		}
	}()
	return messages
}

func sendBindingRequest(conn *net.UDPConn, addr *net.UDPAddr) error {
	m := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	if err := send(m.Raw, conn, addr); err != nil {
		return errors.Wrapf(err, "failed to bind")
	}
	return nil
}

func send(msg []byte, conn *net.UDPConn, addr *net.UDPAddr) error {
	_, err := conn.WriteToUDP(msg, addr)
	if err != nil {
		return errors.Wrapf(err, "failed to send")
	}
	return nil
}
