package sctp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
)

func (c *Client) Handshake(ctx context.Context, addr *net.UDPAddr, timeout time.Duration) error {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	var writeErr error
	var readErr error
	writeDone := false
	send := "syn"
	go func() {
		logger.Infof("Starting handshake write...")
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
		logger.Infof("Stopped handshake write.")
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

	ch := make(chan bool)
	go func() {
		logger.Infof("Wait for handshake...")
		wg.Wait()
		ch <- true
		logger.Infof("Handshake done")
	}()

	select {
	case <-ch:
		break
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(timeout):
		return errors.Errorf("sctp handshake timed out")
	}

	if writeErr != nil {
		return writeErr
	}
	if readErr != nil {
		return readErr
	}

	return nil
}
