package main

import (
	"bytes"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/client"
	log "github.com/sirupsen/logrus"
)

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}

func main() {
	flag.Parse()

	lg := log.New()
	lg.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	client.SetLogger(lg)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))

	urs := "wss://relay.keys.pub/ws"
	cl := client.New(urs)
	cl.Register(alice)

	go func() {
		for {
			msg, err := cl.ReadMessage()
			if err != nil {
				log.Errorf("read err: %v", err)
				time.Sleep(time.Second * 2) // TODO: Backoff
			} else {
				log.Infof("%+v\n", msg)
			}
		}
	}()

	select {
	case <-interrupt:
		cl.Close(true)
	}
}
