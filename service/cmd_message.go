package service

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func messageCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "message",
			Usage: "Message",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "sender, s", Usage: "sender"},
				cli.StringFlag{Name: "recipient, r", Usage: "recipient"},
			},
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "send",
					Usage: "Send",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "sender, s", Usage: "sender"},
						cli.StringFlag{Name: "recipient, r", Usage: "recipient"},
						cli.StringFlag{Name: "message, m", Usage: "message"},
					},
					Action: func(c *cli.Context) error {
						return errors.Errorf("not implemented")
					},
				},
			},
		},
	}
}
