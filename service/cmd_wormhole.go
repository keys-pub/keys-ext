package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

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
							wg.Done()
							return
						}

						if resp.Status == WormholeStatusOpen && !open {
							fmt.Printf("Wormhole open, type a message.\n")
							wg.Done()
						}

						if len(resp.Data) != 0 {
							fmt.Printf("%s\n", string(resp.Data))
						}

						if resp.Status == WormholeStatusClosed {
							fmt.Printf("Wormhole closed.\n")
							return
						}
					}
				}()

				fmt.Printf("Waiting for wormhole to open...\n")
				wg.Wait()

				if recvErr != nil {
					return recvErr
				}

				for {
					reader := bufio.NewReader(os.Stdin)
					text, _ := reader.ReadString('\n')
					if err := client.Send(&WormholeInput{
						Data: []byte(text),
					}); err != nil {
						return err
					}
				}
			},
		},
	}
}
