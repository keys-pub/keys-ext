package service

import (
	"context"
	"fmt"

	"github.com/urfave/cli"
)

func backupCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "backup",
			Usage: "Backup",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "resource, r", Usage: "resource (keyring)"},
			},
			Action: func(c *cli.Context) error {
				resp, err := client.KeysClient().Backup(context.TODO(), &BackupRequest{
					Resource: c.String("resource"),
				})
				if err != nil {
					return err
				}
				// TODO: Make path configurable
				fmt.Println(resp.Path)
				return nil
			},
		},
		cli.Command{
			Name:  "restore",
			Usage: "Restore",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "resource, r", Usage: "resource (keyring)"},
				cli.StringFlag{Name: "path, p", Usage: "path to backup"},
			},
			Action: func(c *cli.Context) error {
				resp, err := client.KeysClient().Restore(context.TODO(), &RestoreRequest{
					Resource: c.String("resource"),
					Path:     c.String("path"),
				})
				if err != nil {
					return err
				}
				printMessage(resp)
				return nil
			},
		},
		cli.Command{
			Name:  "migrate",
			Usage: "Migrate keyring",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "source, s", Usage: "source"},
				cli.StringFlag{Name: "destination, d", Usage: "destination"},
			},
			Hidden: true,
			Action: func(c *cli.Context) error {
				resp, err := client.KeysClient().Migrate(context.TODO(), &MigrateRequest{
					Resource:    c.String("resource"),
					Source:      c.String("source"),
					Destination: c.String("destination"),
				})
				if err != nil {
					return err
				}
				printMessage(resp)
				return nil
			},
		},
	}
}
