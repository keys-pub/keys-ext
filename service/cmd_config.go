package service

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func configCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "config",
			Usage: "Config",
			Subcommands: []cli.Command{
				cli.Command{
					Name:      "set",
					Usage:     "Set config",
					ArgsUsage: "key value",
					Action: func(c *cli.Context) error {
						if c.NArg() != 2 {
							return errors.Errorf("not enough arguments")
						}
						key := c.Args().Get(0)
						value := c.Args().Get(1)

						cfg, err := config(c)
						if err != nil {
							return err
						}
						if !cfg.IsKey(key) {
							return errors.Errorf("unrecognized config key %q", key)
						}
						fmt.Printf("Setting %s=%s\n", key, value)
						cfg.Set(key, value)
						if err := cfg.Save(); err != nil {
							return err
						}

						path, err := cfg.Path(false)
						if err != nil {
							return err
						}
						fmt.Printf("Saved config %s.\n", path)
						fmt.Printf("You should restart the service.\n")
						return nil
					},
				},
			},
			Action: func(c *cli.Context) error {
				cfg, err := config(c)
				if err != nil {
					return err
				}
				b, exportErr := cfg.Export()
				if exportErr != nil {
					return exportErr
				}
				fmt.Printf("%s\n", string(b))
				return nil
			},
		},
	}
}
