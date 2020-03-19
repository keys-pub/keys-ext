package stun

import (
	"net"

	"github.com/pkg/errors"
	"gortc.io/stun"
)

var udp = "udp"

type UDPConn struct {
	peerAddr *net.UDPAddr
	conn     *net.UDPConn
}

func ListenUDP() (*UDPConn, error) {
	conn, err := net.ListenUDP(udp, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to net.ListenUDP")
	}
	c := &UDPConn{conn: conn}
	return c, nil
}

func (c *UDPConn) SendBindingRequest() error {
	srvAddr, err := net.ResolveUDPAddr(udp, stunServer)
	if err != nil {
		return errors.Wrapf(err, "failed to resolve addr")
	}
	if err := sendBindingRequest(c.conn, srvAddr); err != nil {
		return err
	}
	return nil
}

func (c *UDPConn) Send(msg []byte) error {
	if c.peerAddr == nil {
		return errors.Errorf("no peer address set")
	}
	n, err := c.conn.WriteToUDP(msg, c.peerAddr)
	if err != nil {
		return err
	}
	if n != len(msg) {
		return errors.Errorf("failed to (udp) write all bytes")
	}
	return nil
}

func (c *UDPConn) Close() error {
	return c.conn.Close()
}

func (c *UDPConn) Listen() <-chan []byte {
	messages := make(chan []byte)
	go func() {
		for {
			buf := make([]byte, 1024)

			n, _, err := c.conn.ReadFromUDP(buf)
			if err != nil {
				close(messages)
				return
			}
			buf = buf[:n]

			messages <- buf
		}
	}()
	return messages
}

func (c *UDPConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *UDPConn) SetPeer(addr string) error {
	a, err := net.ResolveUDPAddr(udp, addr)
	if err != nil {
		return err
	}
	c.peerAddr = a
	return nil
}

func sendBindingRequest(conn *net.UDPConn, addr *net.UDPAddr) error {
	m := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	if err := sendUDP(m.Raw, conn, addr); err != nil {
		return errors.Wrapf(err, "failed to bind")
	}
	return nil
}

func sendUDP(msg []byte, conn *net.UDPConn, addr *net.UDPAddr) error {
	_, err := conn.WriteToUDP(msg, addr)
	if err != nil {
		return errors.Wrapf(err, "failed to send")
	}
	return nil
}
