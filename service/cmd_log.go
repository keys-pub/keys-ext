package service

import (
	"fmt"

	"github.com/urfave/cli"
)

func logCommands(client *Client, build Build) []cli.Command {
	return []cli.Command{
		{
			Name:   "log",
			Usage:  "Log path",
			Flags:  []cli.Flag{},
			Hidden: true,
			Action: func(c *cli.Context) error {
				env, err := newClientEnv(c, build)
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
