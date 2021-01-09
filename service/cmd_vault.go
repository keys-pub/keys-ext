package service

import (
	"context"
	"fmt"

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
						_, err := client.RPCClient().VaultSync(context.TODO(), &VaultSyncRequest{})
						if err != nil {
							return err
						}
						return nil
					},
				},
				cli.Command{
					Name:  "auth",
					Usage: "Vault auth (single use, expiring)",
					Action: func(c *cli.Context) error {
						resp, err := client.RPCClient().VaultAuth(context.TODO(), &VaultAuthRequest{})
						if err != nil {
							return err
						}
						fmt.Println(resp.Phrase)
						return nil
					},
				},
			},
		},
	}
}
