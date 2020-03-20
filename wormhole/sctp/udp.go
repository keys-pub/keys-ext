package sctp

import (
	"net"

	"github.com/pkg/errors"
	"gortc.io/stun"
)

var udp = "udp"

func stunBindingRequest(conn *net.UDPConn) error {
	srvAddr, err := net.ResolveUDPAddr(udp, stunServer)
	if err != nil {
		return errors.Wrapf(err, "failed to resolve addr")
	}
	if err := sendBindingRequest(conn, srvAddr); err != nil {
		return err
	}
	return nil
}

func listen(conn *net.UDPConn) <-chan []byte {
	messages := make(chan []byte)
	go func() {
		for {
			buf := make([]byte, 1024)

			n, _, err := conn.ReadFromUDP(buf)
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
