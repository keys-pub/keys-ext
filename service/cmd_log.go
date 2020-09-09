package service

import (
	"fmt"

	"github.com/urfave/cli"
)

func logCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "log",
			Usage: "Log path",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				env, err := newClientEnv(c)
				if err != nil {
					return err
				}
				logPath, err := env.LogsPath("keysd.log", false)
				if err != nil {
					return err
				}
				fmt.Println(logPath)
				return nil
			},
		},
	}
}
