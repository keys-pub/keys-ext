package service

import (
	"context"
	"fmt"

	"github.com/urfave/cli"
)

func backupCommands(client *Client) []cli.Command {
	home, _ := homeDir()
	return []cli.Command{
		cli.Command{
			Name:  "backup",
			Usage: "Backup",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "resource, r", Usage: "what to backup", Value: "keyring"},
				cli.StringFlag{Name: "dir, d", Usage: "directory to save to", Value: home},
			},
			Action: func(c *cli.Context) error {
				resp, err := client.KeysClient().Backup(context.TODO(), &BackupRequest{
					Resource: c.String("resource"),
					Dir:      c.String("dir"),
				})
				if err != nil {
					return err
				}
				fmt.Println(resp.Path)
				return nil
			},
		},
		cli.Command{
			Name:  "restore",
			Usage: "Restore",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "resource, r", Usage: "what to restore", Value: "keyring"},
				cli.StringFlag{Name: "path, p", Usage: "path to backup"},
			},
			Action: func(c *cli.Context) error {
				_, err := client.KeysClient().Restore(context.TODO(), &RestoreRequest{
					Resource: c.String("resource"),
					Path:     c.String("path"),
				})
				if err != nil {
					return err
				}
				// printMessage(resp)
				return nil
			},
		},
		cli.Command{
			Name:  "migrate",
			Usage: "Migrate (experimental)",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "resource, r", Usage: "resource"},
				cli.StringFlag{Name: "destination, d", Usage: "destination"},
			},
			Hidden: true,
			Action: func(c *cli.Context) error {
				_, err := client.KeysClient().Migrate(context.TODO(), &MigrateRequest{
					Resource:    c.String("resource"),
					Destination: c.String("destination"),
				})
				if err != nil {
					return err
				}
				// printMessage(resp)
				return nil
			},
		},
	}
}
