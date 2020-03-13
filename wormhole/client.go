package wormhole

import (
	"fmt"
	"net"

	"gortc.io/stun"
)

type Client struct {
	sc   *stun.Client
	addr stun.XORMappedAddress
}

type Addr struct {
	IP   net.IP
	Port int
}

func (a Addr) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

func NewClient() (*Client, error) {
	// Creating a "connection" to STUN server.
	sc, err := stun.Dial("udp", "stun.l.google.com:19302")
	if err != nil {
		return nil, err
	}
	return &Client{
		sc: sc,
	}, nil
}

func (c *Client) Close() error {
	return c.sc.Close()
}

func (c *Client) STUN() error {
	// Building binding request with random transaction id.
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	// Sending request to STUN server, waiting for response message.
	err := c.sc.Do(message, func(res stun.Event) {
		if res.Error != nil {
			logger.Errorf("Stun event error: %v", res.Error)
			return
		}
		// Decoding XOR-MAPPED-ADDRESS attribute from message.
		var addr stun.XORMappedAddress
		if err := addr.GetFrom(res.Message); err != nil {
			panic(err)
		}
		logger.Infof("IP: %s, port: %d", addr.IP, addr.Port)
		c.addr = addr
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Write(addr *Addr) error {
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Call the `Write()` method of the implementor
	// of the `io.Writer` interface.
	_, err = fmt.Fprintf(conn, "hello")
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Listen() error {
	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		return err
	}

	defer func() {
		listener.Close()
		logger.Infof("Listener closed")
	}()

	for {
		_, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}

		// go handleConnection(conn)
	}

	return nil
}
