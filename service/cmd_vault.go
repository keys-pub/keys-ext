package service

import (
	"context"

	"github.com/urfave/cli"
)

func vaultCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "vault",
			Usage: "Vault",
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "sync",
					Usage: "Sync vault",
					Action: func(c *cli.Context) error {
						_, err := client.KeysClient().VaultSync(context.TODO(), &VaultSyncRequest{})
						if err != nil {
							return err
						}
						return nil
					},
				},
			},
		},
	}
}
