package webrtc_test

import (
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/keys-pub/keysd/wormhole/webrtc"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	webrtc.SetLogger(webrtc.NewLogger(webrtc.DebugLevel))

	alice, err := webrtc.NewClient()
	require.NoError(t, err)
	bob, err := webrtc.NewClient()
	require.NoError(t, err)

	messageWg := &sync.WaitGroup{}
	messageWg.Add(2)

	alice.OnMessage(func(message *webrtc.DataChannelMessage) {
		t.Logf("bob: %s", string(message.Data))
		if string(message.Data) == "ping" {
			err := alice.Send([]byte("pong"))
			require.NoError(t, err)
			messageWg.Done()
		}
	})

	bob.OnMessage(func(message *webrtc.DataChannelMessage) {
		t.Logf("alice: %s", string(message.Data))
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

	offer, err := alice.Offer("test")
	require.NoError(t, err)
	answer, err := bob.Answer(offer)
	require.NoError(t, err)
	err = alice.SetAnswer(answer)
	require.NoError(t, err)

	// Wait for channels
	channelWg.Wait()

	err = bob.Send([]byte("ping"))
	require.NoError(t, err)

	t.Logf("Waiting for messages...")
	messageWg.Wait()
	t.Logf("Got messages")

	alice.Close()
	bob.Close()
}

func ExampleNewClient() {
	webrtc.SetLogger(webrtc.NewLogger(webrtc.DebugLevel))

	alice, err := webrtc.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	bob, err := webrtc.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	messageWg := &sync.WaitGroup{}
	messageWg.Add(2)

	alice.OnMessage(func(message *webrtc.DataChannelMessage) {
		fmt.Printf("bob: %s\n", string(message.Data))
		if string(message.Data) == "ping" {
			if err := alice.Send([]byte("pong")); err != nil {
				log.Fatal(err)
			}
			messageWg.Done()
		}
	})

	bob.OnMessage(func(message *webrtc.DataChannelMessage) {
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

	offer, err := alice.Offer("test")
	if err != nil {
		log.Fatal(err)
	}
	answer, err := bob.Answer(offer)
	if err != nil {
		log.Fatal(err)
	}
	if err := alice.SetAnswer(answer); err != nil {
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
