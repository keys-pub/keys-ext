package service

import (
	"time"

	"github.com/keys-pub/keys-ext/ws/client"
)

func (s *service) Relay(srv Keys_RelayServer) error {
	urs := "wss://relay.keys.pub/ws"
	cl, err := client.New(urs)
	if err != nil {
		return err
	}

	// cl.Register(alice)

	for {
		msg, err := cl.ReadMessage()
		if err != nil {
			logger.Errorf("read err: %v", err)
			time.Sleep(time.Second * 2) // TODO: Backoff
		} else {
			logger.Infof("%+v\n", msg)
		}
	}
}
