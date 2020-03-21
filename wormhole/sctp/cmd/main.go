package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
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

	client.OnMessage(func(message []byte) {
		fmt.Printf("Received: %s\n", string(message))
		if string(message) == "ping" {
			fmt.Printf("Sending pong...\n")
			if err := client.Send([]byte("pong")); err != nil {
				log.Fatal(err)
			}
		}
	})

	stunAddr, err := client.STUN(context.TODO(), time.Second*5)
	if err != nil {
		log.Fatal(err)
	}
	addr := &sctp.Addr{IP: stunAddr.IP.String(), Port: stunAddr.Port}
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

	udpAddr, err := net.ResolveUDPAddr("udp", remote.String())
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Handshake(context.TODO(), udpAddr, time.Second*5); err != nil {
		log.Fatal(err)
	}

	if *offer {
		fmt.Printf("Listen...\n")
		if err := client.Listen(context.TODO(), udpAddr); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Connect to %s...\n", udpAddr)
		if err := client.Connect(udpAddr); err != nil {
			log.Fatal(err)
		}

		for {
			fmt.Printf("Sending ping...\n")
			if err := client.Send([]byte("ping")); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second)
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
