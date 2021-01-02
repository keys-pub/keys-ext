package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

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

	tokens := []string{}
	for i := 0; i < 20; i++ {
		tokens = append(tokens, fmt.Sprintf("testtoken%d", i))
	}
	cl.Authorize(tokens)

	<-interrupt
	cl.Close()
}
