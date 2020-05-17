package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func exportCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "export",
			Usage: "Export a key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid"},
				cli.StringFlag{Name: "type, t", Value: "default", Usage: "default, saltpack, ssh"},
				cli.BoolFlag{Name: "public", Usage: "export public part only"},
				cli.StringFlag{Name: "password, p", Usage: "password"},
				cli.BoolFlag{Name: "no-password", Usage: "export without password"},
			},
			Action: func(c *cli.Context) error {
				kid, err := argString(c, "kid", false)
				if err != nil {
					return err
				}

				typ, err := exportTypeFromString(c.String("type"))
				if err != nil {
					return err
				}

				password := c.String("password")
				public := c.Bool("public")
				noPassword := c.Bool("no-password")

				if !public && !noPassword {
					if len(password) == 0 {
						p, err := readPassword("Enter the password:")
						if err != nil {
							return err
						}
						password = p
					}
				}

				req := &KeyExportRequest{
					KID:        kid,
					Password:   password,
					Type:       typ,
					Public:     public,
					NoPassword: noPassword,
				}
				resp, err := client.KeysClient().KeyExport(context.TODO(), req)
				if err != nil {
					return err
				}
				fmt.Println(string(resp.Export))
				return nil
			},
		},
	}
}

func exportTypeFromString(s string) (ExportType, error) {
	switch s {
	case "", "default":
		return DefaultExport, nil
	case "saltpack":
		return SaltpackExport, nil
	case "ssh":
		return SSHExport, nil
	default:
		return DefaultExport, errors.Errorf("invalid type: %s", s)
	}
}
