package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gortc.io/stun"
)

func main() {
	client := NewClient()
	defer client.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	client.OnStunAddr(func(addr string) {
		fmt.Printf("Our address: %s\n", addr)
		wg.Done()
	})

	client.OnMessage(func(message []byte) {
		fmt.Printf("Received: %s\n", string(message))
		if string(message) == "ping" {
			if err := client.Send([]byte("pong")); err != nil {
				log.Fatal(err)
			}
		}
	})

	// Listen
	go func() {
		if err := client.Listen(); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()

	fmt.Printf("Peer address: ")
	addr, err := readAddress()
	if err != nil {
		log.Fatal(err)
	}
	if err := client.SetPeer(addr); err != nil {
		log.Fatal(err)
	}

	if err := client.Send([]byte("ping")); err != nil {
		log.Fatal(err)
	}

	select {}
}

func readAddress() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()
		return text, nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", errors.Errorf("no input")
}

var stunServer = "stun.l.google.com:19302"

type Client struct {
	publicAddr stun.XORMappedAddress
	conn       *udpConn
	onStunAddr func(addr string)
	onMessage  func(message []byte)
}

func NewClient() *Client {
	return &Client{
		onStunAddr: func(addr string) {},
		onMessage:  func([]byte) {},
	}
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) OnStunAddr(f func(addr string)) {
	c.onStunAddr = f
}

func (c *Client) OnMessage(f func(b []byte)) {
	c.onMessage = f
}

// SetPeer sets peer address.
func (c *Client) SetPeer(addr string) error {
	log.Printf("Set peer %s\n", addr)
	return c.conn.SetPeer(addr)
}

// Send to peer.
func (c *Client) Send(message []byte) error {
	return c.conn.Send(message)
}

func (c *Client) Listen() error {
	conn, err := ListenUDP()
	if err != nil {
		return err
	}
	c.conn = conn
	defer c.conn.Close()

	log.Printf("STUN listening on %s\n", c.conn.LocalAddr())

	messageChan := c.conn.Listen()

	if err := c.conn.SendBindingRequest(); err != nil {
		return err
	}

	log.Printf("Waiting for messages...\n")
	for {
		select {
		case <-time.After(time.Second * 10):
			return errors.Errorf("stun timed out")
		case message, ok := <-messageChan:
			if !ok {
				log.Printf("Listen done\n")
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
				log.Printf("Got STUN message: %s\n", xorAddr.String())

				if c.publicAddr.String() != xorAddr.String() {
					log.Printf("Public address: %s\n", xorAddr)
					c.publicAddr = xorAddr
					c.onStunAddr(c.publicAddr.String())
				}
			} else {
				log.Printf("Got message (%d)\n", len(message))
				c.onMessage(message)
			}
		}
	}
}

var udp = "udp"

type udpConn struct {
	peerAddr *net.UDPAddr
	conn     *net.UDPConn
}

func ListenUDP() (*udpConn, error) {
	conn, err := net.ListenUDP(udp, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to net.ListenUDP")
	}
	c := &udpConn{conn: conn}
	return c, nil
}

func (c *udpConn) SendBindingRequest() error {
	srvAddr, err := net.ResolveUDPAddr(udp, stunServer)
	if err != nil {
		return errors.Wrapf(err, "failed to resolve addr")
	}
	if err := sendBindingRequest(c.conn, srvAddr); err != nil {
		return err
	}
	return nil
}

func (c *udpConn) Send(msg []byte) error {
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

func (c *udpConn) Close() error {
	return c.conn.Close()
}

func (c *udpConn) Listen() <-chan []byte {
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

func (c *udpConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *udpConn) SetPeer(addr string) error {
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
