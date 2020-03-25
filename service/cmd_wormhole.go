package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

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
			},
			Action: func(c *cli.Context) error {
				client, err := client.ProtoClient().Wormhole(context.TODO())
				if err != nil {
					return err
				}

				if err := client.Send(&WormholeInput{
					Sender:    c.String("sender"),
					Recipient: c.String("recipient"),
				}); err != nil {
					return err
				}

				open := false
				wg := &sync.WaitGroup{}
				wg.Add(1)

				var recvErr error

				go func() {
					for {
						resp, err := client.Recv()
						if err != nil {
							if err == io.EOF {
								return
							}
							recvErr = err
							if wg != nil {
								wg.Done()
								wg = nil
							}
							return
						}

						if resp.Status == WormholeStatusOpen && !open {
							if wg != nil {
								wg.Done()
								wg = nil
							}
							open = true
						}

						fmtMessage(os.Stdout, resp.Message)

						if resp.Status == WormholeStatusClosed {
							fmt.Printf("Wormhole closed.\n")
							os.Exit(0)
						}
					}
				}()

				fmt.Printf("Waiting for wormhole to open...\n")
				wg.Wait()

				if recvErr != nil {
					return recvErr
				}

				fmt.Printf("Wormhole open, type a message.\n")
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
