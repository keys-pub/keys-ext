package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/client"
)

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))

	urs := "ws://localhost:8080/ws"
	cl := client.New(urs)
	cl.Register(alice)

	go func() {
		for {
			msg, err := cl.ReadMessage()
			if err != nil {
				log.Printf("read err: %v", err)
				time.Sleep(time.Second * 2) // TODO: Backoff
			} else {
				log.Printf("%+v\n", msg)
			}
		}
	}()

	select {
	case <-interrupt:
		cl.Close(true)
	}
}
