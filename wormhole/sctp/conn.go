package sctp

import (
	"net"
	"time"
)

type udpConn struct {
	peerAddr *net.UDPAddr
	conn     *net.UDPConn
}

func (c *udpConn) Read(p []byte) (int, error) {
	n, _, err := c.conn.ReadFromUDP(p)
	if err != nil {
		return 0, err
	}
	return n, err
}

func (c *udpConn) Write(p []byte) (n int, err error) {
	return c.conn.WriteToUDP(p, c.peerAddr)
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
