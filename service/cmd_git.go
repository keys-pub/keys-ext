package service

import (
	"context"
	"io/ioutil"

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
					Name:  "setup",
					Usage: "Setup",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "key, k", Usage: "git ssh key"},
						cli.StringFlag{Name: "url, u", Usage: "git repo url"},
					},
					Action: func(c *cli.Context) error {
						keyFlag := c.String("key")

						var key string
						exists, err := pathExists(keyFlag)
						if err != nil {
							return err
						}
						if exists {
							b, err := ioutil.ReadFile(keyFlag) // #nosec
							if err != nil {
								return err
							}
							key = string(b)
						}

						url := c.String("url")

						req := &GitSetupRequest{
							URL: url,
							Key: key,
						}
						if _, err := client.KeysClient().GitSetup(context.TODO(), req); err != nil {
							return err
						}
						// printResponse(resp)
						return nil
					},
				},
			},
		},
	}
}
