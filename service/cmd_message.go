package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func messagesCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "messages",
			Usage: "List messages",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid"},
			},
			Action: func(c *cli.Context) error {
				resp, err := client.ProtoClient().Messages(context.TODO(), &MessagesRequest{
					KID: c.String("kid"),
				})
				if err != nil {
					return err
				}
				for _, m := range resp.Messages {
					fmt.Printf("%s %s %s\n", m.ID, m.Sender, m.Content.Text)
				}
				return nil
			},
		},
		cli.Command{
			Name:  "message",
			Usage: "Send a message",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid"},
				cli.StringFlag{Name: "sender, s", Usage: "sender (defaults to current key)"},
			},
			ArgsUsage: "<text>",
			Action: func(c *cli.Context) error {
				if c.NArg() > 1 {
					return errors.Errorf("too many arguments")
				}
				resp, err := client.ProtoClient().MessageCreate(context.TODO(), &MessageCreateRequest{
					KID:    c.String("kid"),
					Sender: c.String("sender"),
					Text:   c.Args().First(),
				})
				if err != nil {
					return err
				}
				fmt.Printf("%s\n", resp.Message.ID)
				return nil
			},
		},
	}
}
