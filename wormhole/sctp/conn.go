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

var privateCIDRs = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16"}

func IsPrivateIP(ips string) bool {
	ip := net.ParseIP(ips)
	for _, cidr := range privateCIDRs {
		_, nt, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(err)
		}
		if nt.Contains(ip) {
			return true
		}
	}
	return false
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				// TODO: Support IPv6?
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("no network connection found")
}
