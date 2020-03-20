package ortc

import (
	"net"
	"time"

	"github.com/pkg/errors"
	"gortc.io/stun"
)

func stunAddress() (*stun.XORMappedAddress, *net.UDPConn, error) {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to net.ListenUDP")
	}

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

	stunAddr, err := net.ResolveUDPAddr("udp", "stun.l.google.com:19302")
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to resolve addr")
	}
	if err := sendBindingRequest(conn, stunAddr); err != nil {
		return nil, nil, err
	}

	for {
		select {
		case <-time.After(time.Second * 10):
			return nil, nil, errors.Errorf("stun timed out")
		case message, ok := <-messages:
			if !ok {
				return nil, nil, errors.Errorf("stun connection closed")
			}
			if stun.IsMessage(message) {
				m := new(stun.Message)
				m.Raw = message
				if err := m.Decode(); err != nil {
					return nil, nil, errors.Wrapf(err, "failed to decode stun message")
				}
				var xorAddr stun.XORMappedAddress
				if err := xorAddr.GetFrom(m); err != nil {
					return nil, nil, errors.Wrapf(err, "failed to get address from stun")
				}
				logger.Infof("Got STUN message: %s", xorAddr.String())

				logger.Infof("Public address: %s", xorAddr)
				return &xorAddr, conn, nil

			} else {
				return nil, nil, errors.Errorf("received data on stun address")
			}
		}
	}
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
