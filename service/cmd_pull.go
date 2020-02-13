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
			ArgsUsage: "kid or user (optional)",
			Action: func(c *cli.Context) error {
				identity := c.Args().First()
				req := &PullRequest{
					Identity: identity,
				}
				resp, err := client.ProtoClient().Pull(context.TODO(), req)
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
			ArgsUsage: "kid or user",
			Aliases:   []string{"publish"},
			Action: func(c *cli.Context) error {
				identity := c.Args().First()
				if identity == "" {
					return errors.Errorf("specify kid or user")
				}
				req := &PushRequest{
					Identity: identity,
				}
				resp, err := client.ProtoClient().Push(context.TODO(), req)
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
