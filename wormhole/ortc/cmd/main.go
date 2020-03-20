package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/keys-pub/keysd/wormhole/ortc"
	"github.com/pkg/errors"
)

func main() {
	// stun.SetLogger(stun.NewLogger(stun.DebugLevel))
	offer := flag.Bool("offer", false, "Initiate offer")

	client, err := ortc.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	signal, err := client.Gather()
	if err != nil {
		log.Fatal(err)
	}
	postSignal(signal, *offer)

	remote := readSignal(*offer)

	if err := client.SetRemote(remote, *offer); err != nil {
		log.Fatal(err)
	}
}

func postSignal(signal *ortc.Signal, offer bool) {
	b, err := json.Marshal(signal)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Post signal: %s\n", string(b))
	url := "https://keys.pub/relay/offer"
	if !offer {
		url = "https://keys.pub/relay/answer"
	}
	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	}()
}

func readSignal(offer bool) *ortc.Signal {
	fmt.Printf("Read signal (offer=%t)..\n", offer)
	url := "https://keys.pub/relay/answer"
	if !offer {
		url = "https://keys.pub/relay/offer"
	}
	for {
		fmt.Printf("Get signal..\n")
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode == 200 {
			var signal ortc.Signal
			err = json.NewDecoder(resp.Body).Decode(&signal)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Got signal\n")
			return &signal
		} else if resp.StatusCode == 404 {
			time.Sleep(time.Second)
		} else {
			panic(errors.Errorf("get failed %d", resp.StatusCode))
		}
	}
}
