package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/wormhole/ortc"
	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"
)

func main() {
	// ortc.SetLogger(ortc.NewLogger(ortc.DebugLevel))
	offer := flag.Bool("offer", false, "Initiate offer")
	flag.Parse()

	client, err := ortc.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	signal, err := client.Gather()
	if err != nil {
		log.Fatal(err)
	}
	if *offer {
		if err := writeOffer(signal); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := writeAnswer(signal); err != nil {
			log.Fatal(err)
		}
	}

	var remote *ortc.Signal
	if *offer {
		remote, err = readAnswer()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		remote, err = readOffer()
		if err != nil {
			log.Fatal(err)
		}
	}

	client.OnOpen(func(channel *webrtc.DataChannel) {
		go func() {
			for range time.NewTicker(5 * time.Second).C {
				message := keys.RandPhrase()
				fmt.Printf("Sending: %s\n", message)

				if err := channel.SendText(message); err != nil {
					panic(err)
				}
			}
		}()
	})

	client.OnMessage(func(channel *webrtc.DataChannel, msg webrtc.DataChannelMessage) {
		fmt.Printf("Message (%s): %s\n", channel.Label(), string(msg.Data))
	})

	if err := client.Start(remote, *offer); err != nil {
		log.Fatal(err)
	}

	select {}
}

func writeOffer(offer *ortc.Signal) error {
	return writeSession(offer, "https://keys.pub/relay/offer")
}

func readOffer() (*ortc.Signal, error) {
	return readSession("https://keys.pub/relay/offer")
}

func writeAnswer(answer *ortc.Signal) error {
	return writeSession(answer, "https://keys.pub/relay/answer")
}

func readAnswer() (*ortc.Signal, error) {
	return readSession("https://keys.pub/relay/answer")
}

func writeSession(session *ortc.Signal, url string) error {
	b, err := json.Marshal(session)
	if err != nil {
		return err
	}
	fmt.Printf("Write %s: %s\n", url, string(b))
	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func readSession(url string) (*ortc.Signal, error) {
	for {
		fmt.Printf("Get offer...\n")
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 200 {
			var answer ortc.Signal
			if err = json.NewDecoder(resp.Body).Decode(&answer); err != nil {
				return nil, err
			}
			fmt.Printf("Got offer.\n")
			return &answer, nil
		} else if resp.StatusCode == 404 {
			time.Sleep(time.Second)
		} else {
			return nil, errors.Errorf("Failed to get offer %d", resp.StatusCode)
		}
	}
}
