package wormhole_test

import (
	"fmt"
	"log"
	"sync"

	"github.com/keys-pub/keysd/wormhole"
)

func ExampleNewClient() {
	wormhole.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))

	alice := wormhole.NewClient()
	bob := wormhole.NewClient()

	peerWg := &sync.WaitGroup{}
	peerWg.Add(2)

	alice.SetPublicAddressLn(func(addr string) {
		if err := bob.SetPeer(addr); err != nil {
			log.Fatal(err)
		}
		peerWg.Done()
	})
	bob.SetPublicAddressLn(func(addr string) {
		if err := alice.SetPeer(addr); err != nil {
			log.Fatal(err)
		}
		peerWg.Done()
	})

	messageWg := &sync.WaitGroup{}
	messageWg.Add(2)

	alice.SetMessageLn(func(message []byte) {
		fmt.Printf("bob: %s\n", string(message))
		if string(message) == "ping" {
			if err := alice.Send([]byte("pong")); err != nil {
				log.Fatal(err)
			}
			messageWg.Done()
		}
	})

	bob.SetMessageLn(func(message []byte) {
		fmt.Printf("alice: %s\n", string(message))
		messageWg.Done()
	})

	// Listen
	go func() {
		if err := alice.Listen(); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		if err := bob.Listen(); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for peer addresses
	log.Printf("Wait for peer addresses...\n")
	peerWg.Wait()
	log.Printf("Got peer addresses\n")

	// This message is ignored (needed to allow bob to send)
	if err := alice.Send([]byte("?")); err != nil {
		log.Fatal(err)
	}

	if err := bob.Send([]byte("ping")); err != nil {
		log.Fatal(err)
	}

	log.Printf("Waiting for messages...\n")
	messageWg.Wait()
	log.Printf("Got messages\n")

	alice.Close()
	bob.Close()
	// Output:
	// bob: ping
	// alice: pong
}
