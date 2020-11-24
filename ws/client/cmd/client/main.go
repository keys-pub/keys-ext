package main

import (
	"bytes"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/client"
	"github.com/sirupsen/logrus"
)

var urs = flag.String("url", "wss://relay.keys.pub/ws", "connect using url")

func main() {
	flag.Parse()

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	client.SetLogger(log)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	cl, err := client.New(*urs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			events, err := cl.ReadEvents()
			if err != nil {
				log.Errorf("read err: %v", err)
				time.Sleep(time.Second * 2) // TODO: Backoff
			} else {
				for _, event := range events {
					log.Infof("%+v\n", event)
				}
			}
		}
	}()

	for i := 0; i < 20; i++ {
		key := keys.NewEdX25519KeyFromSeed(testSeed(byte(i)))
		cl.Authorize(key)
	}

	<-interrupt
	cl.Close()
}

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}
