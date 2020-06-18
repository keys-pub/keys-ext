package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	httpclient "github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/wormhole/sctp"
)

func main() {
	sctp.SetLogger(sctp.NewLogger(sctp.DebugLevel))

	offer := flag.Bool("offer", false, "Offer")
	flag.Parse()

	client := sctp.NewClient()
	defer client.Close()

	ctx := context.TODO()

	cmd, err := newCmd()
	if err != nil {
		log.Fatal(err)
	}

	addr, err := client.STUN(ctx, time.Second*5)
	if err != nil {
		log.Fatal(err)
	}
	if *offer {
		if err := cmd.writeOffer(ctx, addr); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := cmd.writeAnswer(ctx, addr); err != nil {
			log.Fatal(err)
		}
	}

	var remote *sctp.Addr
	if *offer {
		remote, err = cmd.readAnswer(ctx)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		remote, err = cmd.readOffer(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *offer {
		fmt.Printf("Connect to %s...\n", remote)
		if err := client.Connect(ctx, remote); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Listen...\n")
		if err := client.ListenForPeer(ctx, remote); err != nil {
			log.Fatal(err)
		}
	}

	if *offer {
		go func() {
			b := make([]byte, 1024)
			for {
				n, err := client.Read(ctx, b)
				if err != nil {
					log.Fatal(err)
				}
				message := b[:n]
				fmt.Printf("Received: %s\n", string(message))
				if string(message) == "answer/ping" {
					fmt.Printf("Sending offer/pong...\n")
					if err := client.Write(ctx, []byte("offer/pong")); err != nil {
						log.Fatal(err)
					}
				}
			}
		}()
		for {
			fmt.Printf("Sending offer/ping...\n")
			if err := client.Write(ctx, []byte("offer/ping")); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second * 5)
		}
	} else {
		go func() {
			b := make([]byte, 1024)
			for {
				n, err := client.Read(ctx, b)
				if err != nil {
					log.Fatal(err)
				}
				message := b[:n]
				fmt.Printf("Received: %s\n", string(message))
				if string(message) == "offer/ping" {
					fmt.Printf("Sending answer/pong...\n")
					if err := client.Write(ctx, []byte("answer/pong")); err != nil {
						log.Fatal(err)
					}
				}
			}
		}()
		for {
			fmt.Printf("Sending answer/ping...\n")
			if err := client.Write(ctx, []byte("answer/ping")); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second * 5)
		}
	}
}

type cmd struct {
	hcl       *httpclient.Client
	offerKey  *keys.EdX25519Key
	answerKey *keys.EdX25519Key
}

func newCmd() (*cmd, error) {
	offerKey := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	answerKey := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	hcl, err := httpclient.New("https://keys.pub")
	if err != nil {
		return nil, err
	}

	return &cmd{
		hcl:       hcl,
		offerKey:  offerKey,
		answerKey: answerKey,
	}, nil
}

func (c *cmd) writeOffer(ctx context.Context, offer *sctp.Addr) error {
	return c.writeSession(ctx, c.offerKey, c.answerKey.ID(), offer, "offer")
}

func (c *cmd) readOffer(ctx context.Context) (*sctp.Addr, error) {
	return c.readSession(ctx, c.offerKey.ID(), c.answerKey, "offer")
}

func (c *cmd) writeAnswer(ctx context.Context, answer *sctp.Addr) error {
	return c.writeSession(ctx, c.answerKey, c.offerKey.ID(), answer, "answer")
}

func (c *cmd) readAnswer(ctx context.Context) (*sctp.Addr, error) {
	return c.readSession(ctx, c.answerKey.ID(), c.offerKey, "answer")
}

func (c *cmd) writeSession(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID, addr *sctp.Addr, typ client.DiscoType) error {
	return c.hcl.DiscoSave(ctx, sender, recipient, typ, addr.String(), time.Minute)
}

func (c *cmd) readSession(ctx context.Context, sender keys.ID, recipient *keys.EdX25519Key, typ client.DiscoType) (*sctp.Addr, error) {
	for {
		fmt.Printf("Get disco (%s)...\n", typ)
		addr, err := c.hcl.Disco(ctx, sender, recipient, typ)
		if err != nil {
			return nil, err
		}
		if addr != "" {
			return sctp.ParseAddr(addr)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
			// Continue
		}
	}
}
