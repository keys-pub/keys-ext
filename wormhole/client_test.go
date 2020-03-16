package wormhole_test

import (
	"fmt"
	"log"
	"sync"

	"github.com/keys-pub/keysd/wormhole"
	"github.com/pion/webrtc/v2"
)

func ExampleNewClient() {
	wormhole.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))

	alice, err := wormhole.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	bob, err := wormhole.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	messageWg := &sync.WaitGroup{}
	messageWg.Add(2)

	alice.OnMessage(func(message webrtc.DataChannelMessage) {
		fmt.Printf("bob: %s\n", string(message.Data))
		if string(message.Data) == "ping" {
			if err := alice.Send([]byte("pong")); err != nil {
				log.Fatal(err)
			}
			messageWg.Done()
		}
	})

	bob.OnMessage(func(message webrtc.DataChannelMessage) {
		fmt.Printf("alice: %s\n", string(message.Data))
		messageWg.Done()
	})

	channelWg := &sync.WaitGroup{}
	channelWg.Add(2)

	alice.OnChannel(func(msg *webrtc.DataChannel) {
		channelWg.Done()
	})

	bob.OnChannel(func(msg *webrtc.DataChannel) {
		channelWg.Done()
	})

	if err := alice.Start(bob.Signal(), true); err != nil {
		log.Fatal(err)
	}

	if err := bob.Start(alice.Signal(), false); err != nil {
		log.Fatal(err)
	}

	// Wait for channels
	channelWg.Wait()

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
