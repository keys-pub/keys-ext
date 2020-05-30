package service

import (
	"context"

	"github.com/urfave/cli"
)

func gitCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:   "git",
			Usage:  "Git",
			Hidden: true,
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "setup",
					Usage: "Setup",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid, k", Usage: "git ssh kid"},
						cli.StringFlag{Name: "url, u", Usage: "git repo url"},
					},
					Action: func(c *cli.Context) error {
						kid := c.String("kid")
						url := c.String("url")
						req := &GitSetupRequest{
							URL: url,
							KID: kid,
						}
						if _, err := client.KeysClient().GitSetup(context.TODO(), req); err != nil {
							return err
						}
						return nil
					},
				},
			},
		},
	}
}
