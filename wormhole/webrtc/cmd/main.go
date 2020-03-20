package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	httpclient "github.com/keys-pub/keysd/http/client"
	"github.com/keys-pub/keysd/wormhole/webrtc"
)

func main() {
	// webrtc.SetLogger(webrtc.NewLogger(webrtc.DebugLevel))
	offer := flag.Bool("offer", false, "Initiate offer")
	generate := flag.Bool("generate", false, "Generate key")
	sender := flag.String("sender", "", "Sender")
	recipient := flag.String("recipient", "", "Recipient")
	flag.Parse()

	kr := newKeyring()
	ks := keys.NewKeystore(kr)

	if *generate {
		generateKey(ks)
		return
	}

	senderID, err := keys.ParseID(*sender)
	if err != nil {
		log.Fatal(err)
	}
	recipientID, err := keys.ParseID(*recipient)
	if err != nil {
		log.Fatal(err)
	}

	hcl, err := httpclient.NewClient("https://keys.pub", ks)
	if err != nil {
		log.Fatal(err)
	}

	client, err := webrtc.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

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
		if err := writeOffer(hcl, offer, senderID, recipientID); err != nil {
			log.Fatal(err)
		}
		answer, err := readAnswer(hcl, senderID, recipientID)
		if err != nil {
			log.Fatal(err)
		}
		if err := client.SetAnswer(answer); err != nil {
			log.Fatal(err)
		}
	} else {
		offer, err := readOffer(hcl, senderID, recipientID)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("Offer:\n")
		// fmt.Printf("%s\n", offer.SDP)
		answer, err := client.Answer(offer)
		if err != nil {
			log.Fatal(err)
		}
		if err := writeAnswer(hcl, answer, senderID, recipientID); err != nil {
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

func newKeyring() keyring.Keyring {
	kr, err := keyring.NewFS("webrtc", os.TempDir())
	if err != nil {
		log.Fatal(err)
	}
	if err := keyring.UnlockWithPassword(kr, "webrtc"); err != nil {
		log.Fatal(err)
	}
	return kr
}

func generateKey(ks *keys.Keystore) {
	key := keys.GenerateEdX25519Key()
	if err := ks.SaveEdX25519Key(key); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated key %s\n", key.ID())
}

func writeOffer(hcl *httpclient.Client, offer *webrtc.SessionDescription, senderID keys.ID, recipientID keys.ID) error {
	fmt.Printf("Write offer...\n")
	b, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	if err := hcl.PutEphemeral(senderID, recipientID, "offer", b); err != nil {
		return err
	}
	return nil
}

func readOffer(hcl *httpclient.Client, senderID keys.ID, recipientID keys.ID) (*webrtc.SessionDescription, error) {
	for {
		fmt.Printf("Read offer...\n")
		ab, err := hcl.GetEphemeral(senderID, recipientID, "offer")
		if err != nil {
			log.Fatal(err)
		}
		if ab != nil {
			var offer webrtc.SessionDescription
			if err := json.Unmarshal(ab, &offer); err != nil {
				log.Fatal(err)
			}
			return &offer, nil
		}
		time.Sleep(time.Second)
	}
}

func writeAnswer(hcl *httpclient.Client, offer *webrtc.SessionDescription, senderID keys.ID, recipientID keys.ID) error {
	fmt.Printf("Write answer...\n")
	b, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	if err := hcl.PutEphemeral(senderID, recipientID, "answer", b); err != nil {
		return err
	}
	return nil
}

func readAnswer(hcl *httpclient.Client, senderID keys.ID, recipientID keys.ID) (*webrtc.SessionDescription, error) {
	for {
		fmt.Printf("Read answer...\n")
		ab, err := hcl.GetEphemeral(senderID, recipientID, "answer")
		if err != nil {
			log.Fatal(err)
		}
		if ab != nil {
			var answer webrtc.SessionDescription
			if err := json.Unmarshal(ab, &answer); err != nil {
				log.Fatal(err)
			}
			return &answer, nil
		}
		time.Sleep(time.Second)
	}
}
