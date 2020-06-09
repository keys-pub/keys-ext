package service

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli"
)

func syncCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:   "sync",
			Usage:  "Sync",
			Hidden: true,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "quiet, q", Usage: "quiet"},
				cli.BoolFlag{Name: "unset", Usage: "unset current program"},
				cli.BoolFlag{Name: "show", Usage: "show current program"},
			},
			Action: func(c *cli.Context) error {
				if c.Bool("unset") {
					_, err := client.KeysClient().SyncUnset(context.TODO(), &SyncUnsetRequest{})
					if err != nil {
						return err
					}
					return nil
				}

				if c.Bool("show") {
					resp, err := client.KeysClient().SyncProgram(context.TODO(), &SyncProgramRequest{})
					if err != nil {
						return err
					}
					if resp.Program != nil {
						printMessage(resp.Program)
					}
					return nil
				}

				name := c.Args().Get(0)
				remote := c.Args().Get(1)

				quiet := c.Bool("quiet")
				syncClient, err := client.KeysClient().Sync(context.TODO(), &SyncRequest{
					Name:   name,
					Remote: remote,
				})
				if err != nil {
					return err
				}

				for {
					resp, recvErr := syncClient.Recv()
					if recvErr != nil {
						if recvErr == io.EOF {
							break
						}
						return recvErr
					}
					if !quiet {
						fmt.Println(resp.Out)
					}
				}

				return nil
			},
		},
	}
}
