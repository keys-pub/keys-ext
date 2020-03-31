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

// ErrHandshakeTimeout if handshake failed.
var ErrHandshakeTimeout = errors.New("sctp handshake timed out")

var stunServer = "stun.l.google.com:19302"

// Client for SCTP.
type Client struct {
	conn *net.UDPConn

	assoc  *sctp.Association
	stream *sctp.Stream
}

// NewClient creates SCTP client.
func NewClient() *Client {
	return &Client{}
}

// Close ...
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

// Write to stream (if connected).
func (c *Client) Write(ctx context.Context, b []byte) error {
	if c.stream == nil {
		return errors.Errorf("no stream")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if _, err := c.stream.Write(b); err != nil {
			return errors.Wrapf(err, "stream write error")
		}
		return nil
	}
}

type read struct {
	N   int
	Err error
}

// Read from stream (if connected).
func (c *Client) Read(ctx context.Context, b []byte) (int, error) {
	if c.stream == nil {
		return 0, errors.Errorf("no stream")
	}
	return c.read(ctx, b)
}

func (c *Client) read(ctx context.Context, b []byte) (int, error) {
	if c.stream == nil {
		return 0, errors.Errorf("no stream")
	}

	readChan := make(chan read)
	go func() {
		// TODO: Use ReadWithContext if that becomes available.
		n, err := c.stream.Read(b)
		readChan <- read{N: n, Err: err}
		close(readChan)
	}()
	select {
	case r := <-readChan:
		return r.N, r.Err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// Local listens for UDP on local address.
// Use STUN instead for a remote address.
func (c *Client) Local() (*Addr, error) {
	conn, err := net.ListenUDP(udp, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to net.ListenUDP")
	}

	// Is there a better way to get the port? (cast to UDPAddr?)
	addr := conn.LocalAddr()
	udpAddr, err := net.ResolveUDPAddr("udp", addr.String())
	if err != nil {
		return nil, err
	}

	local, err := externalIP()
	if err != nil {
		return nil, err
	}
	logger.Infof("Local IP address: %s", local)

	c.conn = conn

	return &Addr{
		IP:   local,
		Port: udpAddr.Port,
	}, nil
}

// STUN initiates the stun requests and returns an address.
func (c *Client) STUN(ctx context.Context, timeout time.Duration) (*Addr, error) {
	if c.conn != nil {
		return nil, errors.Errorf("stun already connected")
	}

	// Ignore local address, we'll get remote address from STUN server
	_, err := c.Local()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to net.ListenUDP")
	}

	logger.Infof("STUN listening on %s", c.conn.LocalAddr())

	messageChan := listen(c.conn)
	// keepAlive := time.NewTicker(time.Second)

	if err := stunBindingRequest(c.conn); err != nil {
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
				// logger.Debugf("STUN: %s", spew.Sdump(message))
				m := new(stun.Message)
				m.Raw = message
				// logger.Debugf("STUN decoded: %+v", m)
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

// Connect to peer.
// Call this with address from STUN.
func (c *Client) Connect(ctx context.Context, peerAddr *Addr) error {
	if c.conn == nil {
		return errors.Errorf("no stun connection, run STUN()")
	}
	if c.assoc != nil {
		return errors.Errorf("client already exists")
	}
	if c.stream != nil {
		return errors.Errorf("stream already exists")
	}

	if err := c.handshake(ctx, peerAddr, true, time.Second*10); err != nil {
		return err
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

// Listen for connections from peer.
func (c *Client) ListenForPeer(ctx context.Context, peerAddr *Addr) error {
	if c.conn == nil {
		return errors.Errorf("no stun connection, run STUN()")
	}
	if c.assoc != nil {
		return errors.Errorf("server already exists")
	}
	if c.stream != nil {
		return errors.Errorf("stream already exists")
	}

	if err := c.handshake(ctx, peerAddr, false, time.Second*10); err != nil {
		return err
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

// Addr is an SCTP address.
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
