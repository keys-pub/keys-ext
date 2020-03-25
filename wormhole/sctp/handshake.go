package sctp

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

func (c *Client) handshake(ctx context.Context, addr *Addr, initiator bool, timeout time.Duration) error {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	udpAddr, err := addr.UDPAddr()
	if err != nil {
		return err
	}

	var writeErr error
	var readErr error
	send := "syn"
	go func() {
		for {
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
				// Established on syn-ack for client
				if initiator {
					break ReadLoop
				}
			case "ack":
				// logger.Debugf("SCTP ack")
				// Established on ack for server
				send = "ack"
				break ReadLoop
			}
		}
		wg.Done()
	}()

	ch := make(chan struct{})
	go func() {
		logger.Debugf("Wait for (sctp) handshake...")
		wg.Wait()
		close(ch)
		logger.Debugf("Handshake (sctp) done")

	}()

	select {
	case <-ch:
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(timeout):
		return errors.Errorf("sctp handshake timed out (%s)", timeout)
	}

	if writeErr != nil {
		return writeErr
	}
	if readErr != nil {
		return readErr
	}

	return nil
}
