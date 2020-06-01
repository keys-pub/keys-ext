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
			Action: func(c *cli.Context) error {
				resp, err := client.KeysClient().Backup(context.TODO(), &BackupRequest{})
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
			Action: func(c *cli.Context) error {
				resp, err := client.KeysClient().Restore(context.TODO(), &RestoreRequest{
					Path: c.Args().First(),
				})
				if err != nil {
					return err
				}
				printResponse(resp)
				return nil
			},
		},
	}
}
