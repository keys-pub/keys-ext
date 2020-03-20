package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/keys-pub/keysd/wormhole/sctp"
	"github.com/pkg/errors"
)

func main() {
	sctp.SetLogger(sctp.NewLogger(sctp.DebugLevel))

	listen := flag.Bool("listen", false, "Listen for connections")
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

	addr, err := client.STUN()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Our address: %s\n", addr)

	fmt.Printf("Peer address: ")
	peerAddr, err := readAddress()
	if err != nil {
		log.Fatal(err)
	}
	udpAddr, err := net.ResolveUDPAddr("udp", peerAddr)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Handshake(udpAddr); err != nil {
		log.Fatal(err)
	}

	if *listen {
		fmt.Printf("Listen...\n")
		if err := client.Listen(udpAddr); err != nil {
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
