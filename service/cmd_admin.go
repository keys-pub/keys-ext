package service

import (
	"context"
	"fmt"

	"github.com/urfave/cli"
)

func adminCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "admin",
			Usage: "Admin",
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "sign-url",
					Usage: "Sign URL",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "signer, s", Usage: "signer"},
						cli.StringFlag{Name: "method, m", Usage: "GET, PUT, POST"},
						cli.StringFlag{Name: "url, u", Usage: "url"},
					},
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
			},
		},
	}
}
