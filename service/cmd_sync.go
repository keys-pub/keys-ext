package service

import (
	"context"

	"github.com/urfave/cli"
)

func syncCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:   "sync",
			Usage:  "Sync keyring",
			Hidden: true,
			Subcommands: []cli.Command{
				syncProgramsCommand(client),
			},
			Action: func(c *cli.Context) error {
				_, err := client.KeysClient().Sync(context.TODO(), &SyncRequest{})
				if err != nil {
					return err
				}
				return nil
			},
		},
	}
}

func syncProgramsCommand(client *Client) cli.Command {
	return cli.Command{
		Name:  "programs",
		Usage: "Programs",
		Subcommands: []cli.Command{
			syncProgramsAddCommand(client),
			syncProgramsRemoveCommand(client),
		},
		Action: func(c *cli.Context) error {
			resp, err := client.KeysClient().SyncPrograms(context.TODO(), &SyncProgramsRequest{})
			if err != nil {
				return err
			}
			printMessage(resp)
			return nil
		},
	}
}

func syncProgramsAddCommand(client *Client) cli.Command {
	return cli.Command{
		Name:  "add",
		Usage: "Add program",
		Action: func(c *cli.Context) error {
			resp, err := client.KeysClient().SyncProgramsAdd(context.TODO(), &SyncProgramsAddRequest{
				Name:   c.Args().Get(0),
				Remote: c.Args().Get(1),
			})
			if err != nil {
				return err
			}
			printMessage(resp)
			return nil
		},
	}
}

func syncProgramsRemoveCommand(client *Client) cli.Command {
	return cli.Command{
		Name:    "remove",
		Aliases: []string{"rm"},
		Usage:   "Remove program",
		Action: func(c *cli.Context) error {
			resp, err := client.KeysClient().SyncProgramsRemove(context.TODO(), &SyncProgramsRemoveRequest{
				ID: c.Args().First(),
			})
			if err != nil {
				return err
			}
			printMessage(resp)
			return nil
		},
	}
}
