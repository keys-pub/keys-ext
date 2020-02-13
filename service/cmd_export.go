package service

import (
	"context"
	"fmt"

	"github.com/urfave/cli"
)

func exportCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "export",
			Usage: "Export a key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid"},
				cli.StringFlag{Name: "password, p", Usage: "password"},
			},
			Action: func(c *cli.Context) error {
				kid, err := argString(c, "kid", false)
				if err != nil {
					return err
				}

				password := c.String("password")
				if len(password) == 0 {
					p, err := readPassword("Enter the password:")
					if err != nil {
						return err
					}
					password = p
				}

				req := &KeyExportRequest{
					KID:      kid,
					Password: password,
					Type:     SaltpackExportType,
				}
				resp, err := client.ProtoClient().KeyExport(context.TODO(), req)
				if err != nil {
					return err
				}
				fmt.Println(string(resp.Export))
				return nil
			},
		},
	}
}
