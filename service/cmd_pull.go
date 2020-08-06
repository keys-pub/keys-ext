package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func pullCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:      "pull",
			Usage:     "Pull from the key server",
			ArgsUsage: "kid or user@service",
			Action: func(c *cli.Context) error {
				key := c.Args().First()
				req := &PullRequest{
					Key: key,
				}
				resp, err := client.KeysClient().Pull(context.TODO(), req)
				if err != nil {
					return err
				}
				for _, kid := range resp.KIDs {
					fmt.Printf("%s\n", kid)
				}
				return nil
			},
		},
		cli.Command{
			Name:      "push",
			Usage:     "Publish to the key server",
			ArgsUsage: "kid or user@service",
			Aliases:   []string{"publish"},
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "check", Usage: "check remote", Hidden: true},
			},
			Action: func(c *cli.Context) error {
				key := c.Args().First()
				if key == "" {
					return errors.Errorf("specify kid or user@service")
				}
				req := &PushRequest{
					Key:         key,
					RemoteCheck: c.Bool("check"),
				}
				resp, err := client.KeysClient().Push(context.TODO(), req)
				if err != nil {
					return err
				}
				for _, url := range resp.URLs {
					fmt.Println(url)
				}
				return nil
			},
		},
	}
}
