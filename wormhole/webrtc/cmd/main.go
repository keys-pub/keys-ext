package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/keys-pub/keysd/wormhole/webrtc"
	"github.com/pkg/errors"
)

func main() {
	// webrtc.SetLogger(webrtc.NewLogger(webrtc.DebugLevel))
	offer := flag.Bool("offer", false, "Initiate offer")
	trace := flag.Bool("trace", false, "Trace (debug)")
	flag.Parse()

	client, err := webrtc.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	client.SetTrace(*trace)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	client.OnStatus(func(status webrtc.Status) {
		switch status {
		case webrtc.Disconnected:
			log.Printf("Disconnected.\n")
			os.Exit(2)
		case webrtc.Closed:
			log.Printf("Closed.\n")
			os.Exit(1)
		default:
			log.Printf("Status: %s\n", status)
		}

	})
	client.OnOpen(func(channel webrtc.Channel) {
		fmt.Printf("opened\n")
		wg.Done()
	})
	client.OnMessage(func(channel webrtc.Channel, message webrtc.Message) {
		fmt.Printf("Recieved: %s\n", string(message.Data()))
		if string(message.Data()) == "ping" {
			fmt.Printf("Send pong...\n")
			if err := channel.Send([]byte("pong")); err != nil {
				panic(err)
			}
		}
	})
	client.OnClose(func(channel webrtc.Channel) {
		fmt.Printf("closed\n")
	})

	if *offer {
		offer, err := client.Offer()
		if err != nil {
			log.Fatal(err)
		}
		if err := writeOffer(offer); err != nil {
			log.Fatal(err)
		}
		answer, err := readAnswer()
		if err != nil {
			log.Fatal(err)
		}
		if err := client.SetAnswer(answer); err != nil {
			log.Fatal(err)
		}
	} else {
		offer, err := readOffer()
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("Offer:\n")
		// fmt.Printf("%s\n", offer.SDP)
		answer, err := client.Answer(offer)
		if err != nil {
			log.Fatal(err)
		}
		if err := writeAnswer(answer); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("Waiting for channel...\n")
	wg.Wait()

	fmt.Printf("Send ping...\n")
	if err := client.Send([]byte("ping")); err != nil {
		log.Fatal(err)
	}

	select {}
}

func writeOffer(offer *webrtc.SessionDescription) error {
	b, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	fmt.Printf("Write offer: %s\n", string(b))
	resp, err := http.Post("https://keys.pub/relay/offer", "application/json; charset=utf-8", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func readOffer() (*webrtc.SessionDescription, error) {
	for {
		fmt.Printf("Get offer...\n")
		resp, err := http.Get("https://keys.pub/relay/offer")
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 200 {
			var answer webrtc.SessionDescription
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

func writeAnswer(answer *webrtc.SessionDescription) error {
	b, err := json.Marshal(answer)
	if err != nil {
		return err
	}
	fmt.Printf("Write answer: %s\n", string(b))
	resp, err := http.Post("https://keys.pub/relay/answer", "application/json; charset=utf-8", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func readAnswer() (*webrtc.SessionDescription, error) {
	for {
		fmt.Printf("Get answer...\n")
		resp, err := http.Get("https://keys.pub/relay/answer")
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 200 {
			var answer webrtc.SessionDescription
			if err = json.NewDecoder(resp.Body).Decode(&answer); err != nil {
				return nil, err
			}
			fmt.Printf("Got answer.\n")
			return &answer, nil
		} else if resp.StatusCode == 404 {
			time.Sleep(time.Second)
		} else {
			return nil, errors.Errorf("Failed to get offer %d", resp.StatusCode)
		}
	}
}