package sctp

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func (c *Client) Handshake(addr *net.UDPAddr) error {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	var writeErr error
	var readErr error
	writeDone := false
	send := "syn"
	go func() {
		logger.Infof("Sending syn/ack...")
		for !writeDone {
			if _, err := c.conn.WriteToUDP([]byte(send), addr); err != nil {
				writeErr = err
				writeDone = true
			}
			if send == "ack" {
				writeDone = true
			}
			if !writeDone {
				time.Sleep(time.Millisecond * 100)
			}
		}
		logger.Infof("Stopped sending syn/ack")
		wg.Done()
	}()

	readDone := false
	go func() {
		for !readDone {
			b, err := c.readFromUDP()
			if err != nil {
				readErr = err
				break
			}
			switch string(b) {
			case "syn":
				fmt.Printf("Received syn.\n")
				send = "syn-ack"
			case "syn-ack":
				fmt.Printf("Received syn-ack.\n")
				send = "ack"
			case "ack":
				fmt.Printf("Received ack.\n")
				send = "ack"
				readDone = true
			}
		}
		wg.Done()
	}()

	wg.Wait()

	if writeErr != nil {
		return writeErr
	}
	if readErr != nil {
		return readErr
	}

	return nil
}
