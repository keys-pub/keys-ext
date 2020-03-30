package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/keys-pub/keysd/wormhole"
	"github.com/urfave/cli"
)

func wormholeCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "wormhole",
			Usage: "Wormhole",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "sender, s", Usage: "sender"},
				cli.StringFlag{Name: "recipient, r", Usage: "recipient"},
				cli.StringFlag{Name: "invite", Usage: "invite code"},
			},
			Action: func(c *cli.Context) error {
				client, err := client.ProtoClient().Wormhole(context.TODO())
				if err != nil {
					return err
				}

				fmt.Printf("Starting wormhole...\n")

				if err := client.Send(&WormholeInput{
					Sender:    c.String("sender"),
					Recipient: c.String("recipient"),
					Invite:    c.String("invite"),
				}); err != nil {
					return err
				}

				var status WormholeStatus

				go func() {
					for {
						resp, err := client.Recv()
						if err != nil {
							if err == io.EOF {
								os.Exit(0)
								return
							}
							clientFatal(err)
						}

						if resp.Status != status {
							status = resp.Status
							switch status {
							case WormholeStarting:
							case WormholeOffering:
								// fmt.Printf("Offering...\n")
							case WormholeAnswering:
								// fmt.Printf("Found offer, answering...\n")
							case WormholeHandshake:
								fmt.Printf("Trying handshake...\n")
							case WormholeConnected:
								fmt.Printf("Wormhole connected, type a message.\n")
							case WormholeClosed:
								fmt.Printf("Wormhole closed.\n")
								go func() {
									// TODO: Get error before close status so we don't have to sleep
									time.Sleep(time.Second)
									os.Exit(0)
								}()
							}
						}

						fmtMessage(os.Stdout, resp.Message)
					}
				}()

				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					text := scanner.Text()
					id := wormhole.NewID()
					if err := client.Send(&WormholeInput{
						ID:   id,
						Data: []byte(text),
						Type: UTF8Content,
					}); err != nil {
						return err
					}
				}
				if err := scanner.Err(); err != nil {
					return err
				}

				return nil
			},
		},
	}
}
