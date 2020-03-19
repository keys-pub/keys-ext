package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keysd/wormhole/webrtc"
	"github.com/pkg/errors"
)

func main() {
	// webrtc.SetLogger(webrtc.NewLogger(webrtc.DebugLevel))
	offer := flag.Bool("offer", false, "Initiate offer")
	flag.Parse()

	client, err := webrtc.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	client.OnOpen(func(channel webrtc.Channel) {
		fmt.Printf("opened\n")
		wg.Done()
	})
	client.OnMessage(func(message webrtc.Message) {
		fmt.Printf("Recieved: %s\n", string(message.Data()))
		if string(message.Data()) == "ping" {
			if err := client.Send([]byte("pong")); err != nil {
				panic(err)
			}
		}
	})
	client.OnClose(func(channel webrtc.Channel) {
		fmt.Printf("closed\n")
	})

	if *offer {
		offer, err := client.Offer("test")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Offer:\n")
		if err := writeSession(offer); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Enter answer:\n")
		answer, err := readSession()
		if err != nil {
			log.Fatal(err)
		}
		if err := client.SetAnswer(answer); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Enter offer:\n")
		offer, err := readSession()
		if err != nil {
			log.Fatal(err)
		}
		answer, err := client.Answer(offer)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Answer:\n")
		if err := writeSession(answer); err != nil {
			log.Fatal(err)
		}
	}

	wg.Wait()

	fmt.Printf("Send ping...\n")
	if err := client.Send([]byte("ping")); err != nil {
		log.Fatal(err)
	}

	select {}
}

func readSession() (*webrtc.SessionDescription, error) {
	scanner := bufio.NewScanner(os.Stdin)
	input := ""

	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			input = input + strings.TrimSpace(text)
		} else {
			dec, err := encoding.Decode(input, encoding.Base64)
			if err != nil {
				return nil, err
			}

			r, err := gzip.NewReader(bytes.NewBuffer(dec))
			if err != nil {
				return nil, err
			}
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				return nil, err
			}

			var session webrtc.SessionDescription
			if err := json.Unmarshal(buf.Bytes(), &session); err != nil {
				log.Fatal(err)
			}
			return &session, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, errors.Errorf("no input")
}

func writeSession(s *webrtc.SessionDescription) error {
	mb, err := json.Marshal(s)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(mb); err != nil {
		return err
	}
	gz.Flush()
	gz.Close()
	enc, err := encoding.Encode(b.Bytes(), encoding.Base64)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", enc)
	return nil
}
