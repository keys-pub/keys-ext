package service

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func envCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "env",
			Usage: "Env",
			Subcommands: []cli.Command{
				cli.Command{
					Name:      "set",
					Usage:     "Set env value",
					ArgsUsage: "key value",
					Action: func(c *cli.Context) error {
						if c.NArg() != 2 {
							return errors.Errorf("not enough arguments")
						}
						key := c.Args().Get(0)
						value := c.Args().Get(1)

						env, err := newClientEnv(c)
						if err != nil {
							return err
						}
						if !env.IsKey(key) {
							return errors.Errorf("unrecognized env key %q", key)
						}
						fmt.Printf("Setting %s=%s\n", key, value)
						env.Set(key, value)
						if err := env.Save(); err != nil {
							return err
						}

						path, err := env.Path(false)
						if err != nil {
							return err
						}
						fmt.Printf("Saved env %s.\n", path)
						fmt.Printf("You should restart the service.\n")
						return nil
					},
				},
			},
			Action: func(c *cli.Context) error {
				env, err := newClientEnv(c)
				if err != nil {
					return err
				}
				b, err := env.Export()
				if err != nil {
					return err
				}
				fmt.Printf("%s\n", string(b))
				return nil
			},
		},
	}
}
