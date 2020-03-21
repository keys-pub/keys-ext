package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/keys-pub/keysd/wormhole/sctp"
	"github.com/pkg/errors"
)

func main() {
	sctp.SetLogger(sctp.NewLogger(sctp.DebugLevel))

	offer := flag.Bool("offer", false, "Offer")
	flag.Parse()

	client := sctp.NewClient()
	defer client.Close()

	addr, err := client.STUN(context.TODO(), time.Second*5)
	if err != nil {
		log.Fatal(err)
	}
	if *offer {
		if err := writeOffer(addr); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := writeAnswer(addr); err != nil {
			log.Fatal(err)
		}
	}

	var remote *sctp.Addr
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

	if err := client.Handshake(context.TODO(), remote, time.Second*5); err != nil {
		log.Fatal(err)
	}

	if *offer {
		fmt.Printf("Connect to %s...\n", remote)
		if err := client.Connect(remote); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Listen...\n")
		if err := client.Listen(context.TODO(), remote); err != nil {
			log.Fatal(err)
		}
	}

	if *offer {
		go func() {
			b := make([]byte, 1024)
			for {
				n, err := client.Read(b)
				if err != nil {
					log.Fatal(err)
				}
				message := b[:n]
				fmt.Printf("Received: %s\n", string(message))
				if string(message) == "answer/ping" {
					fmt.Printf("Sending offer/pong...\n")
					if err := client.Write([]byte("offer/pong")); err != nil {
						log.Fatal(err)
					}
				}
			}
		}()
		for {
			fmt.Printf("Sending offer/ping...\n")
			if err := client.Write([]byte("offer/ping")); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second * 5)
		}
	} else {
		go func() {
			b := make([]byte, 1024)
			for {
				n, err := client.Read(b)
				if err != nil {
					log.Fatal(err)
				}
				message := b[:n]
				fmt.Printf("Received: %s\n", string(message))
				if string(message) == "offer/ping" {
					fmt.Printf("Sending answer/pong...\n")
					if err := client.Write([]byte("answer/pong")); err != nil {
						log.Fatal(err)
					}
				}
			}
		}()
		for {
			fmt.Printf("Sending answer/ping...\n")
			if err := client.Write([]byte("answer/ping")); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func writeOffer(offer *sctp.Addr) error {
	return writeSession(offer, "https://keys.pub/relay/offer")
}

func readOffer() (*sctp.Addr, error) {
	return readSession("https://keys.pub/relay/offer")
}

func writeAnswer(answer *sctp.Addr) error {
	return writeSession(answer, "https://keys.pub/relay/answer")
}

func readAnswer() (*sctp.Addr, error) {
	return readSession("https://keys.pub/relay/answer")
}

func writeSession(session *sctp.Addr, url string) error {
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

func readSession(url string) (*sctp.Addr, error) {
	for {
		fmt.Printf("Get offer...\n")
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 200 {
			var answer sctp.Addr
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
