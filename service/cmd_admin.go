package service

import (
	"context"
	"fmt"

	"github.com/urfave/cli"
)

func adminCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:   "admin",
			Usage:  "Admin",
			Hidden: true,
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "sign-url",
					Usage: "Sign URL",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "signer, s", Usage: "signer"},
						cli.StringFlag{Name: "method, m", Usage: "GET, PUT, POST"},
						cli.StringFlag{Name: "url, u", Usage: "url"},
					},
					Hidden:    true,
					ArgsUsage: "",
					Action: func(c *cli.Context) error {
						req := &AdminSignURLRequest{
							Signer: c.String("signer"),
							Method: c.String("method"),
							URL:    c.String("url"),
						}
						resp, err := client.KeysClient().AdminSignURL(context.TODO(), req)
						if err != nil {
							return err
						}
						fmt.Printf("%+v\n", resp)
						return nil
					},
				},
				cli.Command{
					Name:   "check",
					Usage:  "Check key",
					Hidden: true,
					Flags: []cli.Flag{
						cli.StringFlag{Name: "signer, s", Usage: "signer"},
						cli.StringFlag{Name: "kid, k", Usage: "key to update"},
					},
					Action: func(c *cli.Context) error {
						req := &AdminCheckRequest{
							Signer: c.String("signer"),
							KID:    c.String("kid"),
						}
						resp, err := client.KeysClient().AdminCheck(context.TODO(), req)
						if err != nil {
							return err
						}
						fmt.Printf("%+v\n", resp)
						return nil
					},
				},
			},
		},
	}
}
