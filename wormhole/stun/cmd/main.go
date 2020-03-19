package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/keys-pub/keysd/wormhole/stun"
	"github.com/pkg/errors"
)

func main() {
	// stun.SetLogger(stun.NewLogger(stun.DebugLevel))

	client := stun.NewClient()
	defer client.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	client.OnPeer(func(addr string) {
		fmt.Printf("Our address: %s\n", addr)
		wg.Done()
	})

	client.OnMessage(func(message []byte) {
		fmt.Printf("Received: %s\n", string(message))
		if string(message) == "ping" {
			if err := client.Send([]byte("pong")); err != nil {
				log.Fatal(err)
			}
		}
	})

	// Listen
	go func() {
		if err := client.Listen(); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()

	fmt.Printf("Peer address: ")
	addr, err := readAddress()
	if err != nil {
		log.Fatal(err)
	}
	if err := client.SetPeer(addr); err != nil {
		log.Fatal(err)
	}

	if err := client.Send([]byte("ping")); err != nil {
		log.Fatal(err)
	}

	select {}
}

func readAddress() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()
		return text, nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", errors.Errorf("no input")
}
