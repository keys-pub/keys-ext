package service

import (
	"context"

	"github.com/urfave/cli"
)

func gitCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:   "git",
			Usage:  "Git",
			Hidden: true,
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "import",
					Usage: "Import into a git repository",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "key", Usage: "git ssh key path", Value: homePath(".ssh", "id_ed25519")},
						cli.StringFlag{Name: "url, u", Usage: "git repo url"},
					},
					Action: func(c *cli.Context) error {
						req := &GitImportRequest{
							URL:     c.String("url"),
							KeyPath: c.String("key"),
						}
						if _, err := client.KeysClient().GitImport(context.TODO(), req); err != nil {
							return err
						}
						return nil
					},
				},
				cli.Command{
					Name:  "clone",
					Usage: "Clone a git repository",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "key", Usage: "git ssh key path", Value: homePath(".ssh", "id_ed25519")},
						cli.StringFlag{Name: "url, u", Usage: "git repo url"},
					},
					Action: func(c *cli.Context) error {
						req := &GitCloneRequest{
							URL:     c.String("url"),
							KeyPath: c.String("key"),
						}
						if _, err := client.KeysClient().GitClone(context.TODO(), req); err != nil {
							return err
						}
						return nil
					},
				},
			},
		},
	}
}
