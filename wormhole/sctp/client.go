package sctp

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/pion/logging"
	"github.com/pion/sctp"
	"github.com/pkg/errors"
	"gortc.io/stun"
)

var stunServer = "stun.l.google.com:19302"

type Client struct {
	conn      *net.UDPConn
	onMessage func(message []byte)

	assoc  *sctp.Association
	stream *sctp.Stream
}

func NewClient() *Client {
	return &Client{
		onMessage: func([]byte) {},
	}
}

func (c *Client) Close() {
	if c.stream != nil {
		c.stream.Close()
	}
	if c.assoc != nil {
		c.assoc.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) OnMessage(f func(b []byte)) {
	c.onMessage = f
}

func (c *Client) Connect(peerAddr *net.UDPAddr) error {
	if c.conn == nil {
		return errors.Errorf("no stun connection, run STUN()")
	}

	config := sctp.Config{
		NetConn:       &udpConn{conn: c.conn, peerAddr: peerAddr},
		LoggerFactory: logging.NewDefaultLoggerFactory(),
	}
	logger.Infof("Create client association...")
	a, err := sctp.Client(config)
	if err != nil {
		return err
	}
	c.assoc = a

	logger.Infof("Open stream read...")
	stream, err := a.OpenStream(0, sctp.PayloadTypeWebRTCDCEP)
	if err != nil {
		return err
	}
	logger.Infof("Stream opened.")
	stream.SetReliabilityParams(false, sctp.ReliabilityTypeReliable, 10)
	c.stream = stream

	go func() {
		logger.Infof("Stream read.")
		if err := c.read(); err != nil {
			// TODO: Notify read err
			logger.Errorf("Read error: %v", err)
		}
	}()

	return nil
}

func (c *Client) read() error {
	buf := make([]byte, 1024)
	for {
		n, err := c.stream.Read(buf)
		if err != nil {
			return err
		}
		c.onMessage(buf[:n])
	}
}

func (c *Client) Send(message []byte) error {
	if c.stream == nil {
		return errors.Errorf("no stream")
	}
	if _, err := c.stream.Write(message); err != nil {
		return err
	}
	return nil
}

func (c *Client) STUN() (*stun.XORMappedAddress, error) {
	if c.conn != nil {
		return nil, errors.Errorf("stun already connected")
	}
	conn, err := net.ListenUDP(udp, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to net.ListenUDP")
	}
	c.conn = conn

	logger.Infof("Listening on %s", conn.LocalAddr())

	messageChan := listen(conn)
	// keepAlive := time.NewTicker(time.Second)

	if err := stunBindingRequest(conn); err != nil {
		return nil, err
	}

	logger.Infof("Waiting for stun...")
	var stunAddr stun.XORMappedAddress
Stun:
	for {
		select {
		case <-time.After(time.Second * 10):
			return nil, errors.Errorf("stun timed out")
		case message, ok := <-messageChan:
			if !ok {
				return nil, errors.Errorf("stun connection closed")
			}
			if stun.IsMessage(message) {
				m := new(stun.Message)
				m.Raw = message
				if err := m.Decode(); err != nil {
					return nil, errors.Wrapf(err, "failed to decode stun message")
				}
				var xorAddr stun.XORMappedAddress
				if err := xorAddr.GetFrom(m); err != nil {
					return nil, errors.Wrapf(err, "failed to get address from stun")
				}
				logger.Infof("Stun address: %s", xorAddr)
				stunAddr = xorAddr
				break Stun
			}
		}
	}

	return &stunAddr, nil
}

func (c *Client) Listen(peerAddr *net.UDPAddr) error {
	if c.conn == nil {
		return errors.Errorf("no stun connection, run STUN()")
	}

	slog := logging.NewDefaultLoggerFactory()
	// slog.DefaultLogLevel = logging.LogLevelTrace
	slog.Writer = os.Stderr
	config := sctp.Config{
		NetConn:       &udpConn{conn: c.conn, peerAddr: peerAddr},
		LoggerFactory: slog,
	}
	logger.Infof("Create server association...")
	a, err := sctp.Server(config)
	if err != nil {
		return err
	}
	c.assoc = a

	logger.Infof("Accept stream...")
	stream, err := a.AcceptStream()
	if err != nil {
		log.Fatal(err)
	}
	logger.Infof("Stream accepted.")
	stream.SetReliabilityParams(false, sctp.ReliabilityTypeReliable, 10)
	c.stream = stream

	logger.Infof("Stream read...")
	return c.read()
}

func (c *Client) writeToUDP(b []byte, addr string) error {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	if _, err := c.conn.WriteToUDP(b, a); err != nil {
		return err
	}
	return nil
}

func (c *Client) readFromUDP() ([]byte, error) {
	buf := make([]byte, 1024)
	n, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]
	return buf, nil
}
