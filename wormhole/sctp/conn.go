package sctp

import (
	"net"
	"time"

	"github.com/pkg/errors"
)

type udpConn struct {
	peerAddr *net.UDPAddr
	conn     *net.UDPConn
}

func newUDPConn(conn *net.UDPConn, addr *Addr) (*udpConn, error) {
	peerAddr, err := addr.UDPAddr()
	if err != nil {
		return nil, err
	}
	return &udpConn{
		peerAddr: peerAddr,
		conn:     conn,
	}, nil
}

func (c *udpConn) Read(p []byte) (int, error) {
	n, addr, err := c.conn.ReadFromUDP(p)
	if err != nil {
		return 0, errors.Wrapf(err, "udp read error")
	}
	if addr.String() != c.peerAddr.String() {
		return 0, errors.Errorf("received data from an unexpected address")
	}
	return n, nil
}

func (c *udpConn) Write(p []byte) (int, error) {
	n, err := c.conn.WriteToUDP(p, c.peerAddr)
	if err != nil {
		return n, errors.Wrapf(err, "udp write error")
	}
	return n, nil
}

func (c *udpConn) Close() error {
	return c.conn.Close()
}

func (c *udpConn) LocalAddr() net.Addr {
	if c.conn != nil {
		return c.conn.LocalAddr()
	}
	return nil
}

func (c *udpConn) RemoteAddr() net.Addr {
	return c.peerAddr
}

func (c *udpConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *udpConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *udpConn) SetWriteDeadline(t time.Time) error {
	return nil
}
