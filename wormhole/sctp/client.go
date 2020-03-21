package sctp

import (
	"context"
	"fmt"
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
	conn *net.UDPConn

	assoc  *sctp.Association
	stream *sctp.Stream
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Close() {
	c.close(false)
}

func (c *Client) close(notify bool) {
	// TODO: notify

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

func (c *Client) Write(b []byte) error {
	if c.stream == nil {
		return errors.Errorf("no stream")
	}
	if _, err := c.stream.Write(b); err != nil {
		return err
	}
	return nil
}

func (c *Client) Read(b []byte) (int, error) {
	return c.stream.Read(b)
}

func (c *Client) STUN(ctx context.Context, timeout time.Duration) (*Addr, error) {
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
		case <-ctx.Done():
			return nil, ctx.Err()
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

	return &Addr{
		IP:   stunAddr.IP.String(),
		Port: stunAddr.Port,
	}, nil
}

func (c *Client) Connect(peerAddr *Addr) error {
	if c.conn == nil {
		return errors.Errorf("no stun connection, run STUN()")
	}
	if c.assoc != nil {
		return errors.Errorf("client already exists")
	}
	if c.stream != nil {
		return errors.Errorf("stream already exists")
	}

	netConn, err := newUDPConn(c.conn, peerAddr)
	if err != nil {
		return err
	}
	config := sctp.Config{
		NetConn:       netConn,
		LoggerFactory: logging.NewDefaultLoggerFactory(),
	}
	logger.Infof("Create client (for peer %s)...", peerAddr)
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
	stream.SetReliabilityParams(false, sctp.ReliabilityTypeReliable, 0)
	c.stream = stream

	return nil
}

func (c *Client) Listen(ctx context.Context, peerAddr *Addr) error {
	if c.conn == nil {
		return errors.Errorf("no stun connection, run STUN()")
	}
	if c.assoc != nil {
		return errors.Errorf("server already exists")
	}
	if c.stream != nil {
		return errors.Errorf("stream already exists")
	}

	slog := logging.NewDefaultLoggerFactory()
	// slog.DefaultLogLevel = logging.LogLevelTrace
	slog.Writer = os.Stderr

	netConn, err := newUDPConn(c.conn, peerAddr)
	if err != nil {
		return err
	}
	config := sctp.Config{
		NetConn:       netConn,
		LoggerFactory: slog,
	}
	logger.Infof("Create server (for peer %s)...", peerAddr)
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
	stream.SetReliabilityParams(false, sctp.ReliabilityTypeReliable, 0)
	c.stream = stream

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

type Addr struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

func (a Addr) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

func (a Addr) UDPAddr() (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", a.String())

}
