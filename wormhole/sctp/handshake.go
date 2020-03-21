package sctp

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

func (c *Client) handshake(ctx context.Context, addr *Addr, timeout time.Duration) error {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	udpAddr, err := addr.UDPAddr()
	if err != nil {
		return err
	}

	var writeErr error
	var readErr error
	write := true
	send := "syn"
	go func() {
		for write {
			if _, err := c.conn.WriteToUDP([]byte(send), udpAddr); err != nil {
				writeErr = err
				break
			}
			if send == "ack" {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}
		wg.Done()
	}()

	read := true
	go func() {
	ReadLoop:
		for read {
			b, err := c.readFromUDP()
			if err != nil {
				readErr = err
				break
			}
			switch string(b) {
			case "syn":
				// logger.Debugf("SCTP syn")
				send = "syn-ack"
			case "syn-ack":
				// logger.Debugf("SCTP syn-ack")
				send = "ack"
			case "ack":
				// logger.Debugf("SCTP ack")
				send = "ack"
				break ReadLoop
			}
		}
		wg.Done()
	}()

	ch := make(chan bool)
	go func() {
		logger.Debugf("Wait for (sctp) handshake...")
		wg.Wait()
		ch <- true
		logger.Debugf("Handshake (sctp) done")
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
