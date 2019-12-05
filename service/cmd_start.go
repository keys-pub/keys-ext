package service

import (
	"fmt"

	"github.com/urfave/cli"
)

func startCommands() []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "restart",
			Usage: "Restart the service",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				cfg, err := config(c)
				if err != nil {
					return err
				}
				restartErr := restart(cfg)
				if restartErr != nil {
					return restartErr
				}
				fmt.Printf("Restarted.\n")
				return nil
			},
		},
		cli.Command{
			Name:  "start",
			Usage: "Start the service",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				cfg, err := config(c)
				if err != nil {
					return err
				}
				if err := start(cfg, true); err != nil {
					return err
				}
				fmt.Println("Started.")
				return nil
			},
		},
		cli.Command{
			Name:  "stop",
			Usage: "Stop the service",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				cfg, err := config(c)
				if err != nil {
					return err
				}
				if err := stop(cfg); err != nil {
					return err
				}
				fmt.Println("Stopped.")
				return nil
			},
		},
		cli.Command{
			Name:  "uninstall",
			Usage: "Uninstall",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				cfg, err := config(c)
				if err != nil {
					return err
				}
				return Uninstall(cfg)
			},
		},
	}
}
