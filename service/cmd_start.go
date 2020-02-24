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
				fmt.Println("restarted")
				return nil
			},
		},
		cli.Command{
			Name:  "start",
			Usage: "Start the service",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "from", Usage: "where we are being started from", Hidden: true},
			},
			Action: func(c *cli.Context) error {
				cfg, err := config(c)
				if err != nil {
					return err
				}
				// If we start from the app...
				if c.String("from") == "app" {
					if err := startFromApp(cfg); err != nil {
						return err
					}
				} else {
					if err := start(cfg, true); err != nil {
						return err
					}
				}
				fmt.Println("started")
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
				fmt.Println("stopped")
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
