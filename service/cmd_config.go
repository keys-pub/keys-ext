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
						// _, err := client.ProtoClient().ConfigSet(context.TODO(), &ConfigSetRequest{
						// 	Key:   key,
						// 	Value: value,
						// })
						// if err != nil {
						// 	return err
						// }

						cfg, err := config(c)
						if err != nil {
							return err
						}
						cfg.Set(key, value)
						if err := cfg.Save(); err != nil {
							return err
						}

						fmt.Printf("Saved config.\n")

						// Stop after config change (if running)
						if err := stop(cfg); err != nil {
							if errors.Cause(err) != errNotRunning {
								return err
							}
						} else {
							fmt.Printf("Service stopped.\n")
						}
						return nil
					},
				},
			},
			Action: func(c *cli.Context) error {
				// configResp, configErr := client.ProtoClient().Config(context.TODO(), &ConfigRequest{})
				// if configErr != nil {
				// 	return configErr
				// }
				// for k, v := range configResp.Config {
				// 	fmt.Printf("%q: %q", k, v)
				// }
				// return nil
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
