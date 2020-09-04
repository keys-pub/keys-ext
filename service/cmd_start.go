package service

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/urfave/cli"
)

func startCommands() []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "restart",
			Usage: "Restart the service",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				env, err := newClientEnv(c)
				if err != nil {
					return err
				}
				restartErr := restart(env)
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
				cli.IntFlag{Name: "port", Usage: "port", Hidden: true},
			},
			Action: func(c *cli.Context) error {
				env, err := newClientEnv(c)
				if err != nil {
					return err
				}
				port := c.Int("port")
				if port != 0 {
					if err := env.savePortFlag(port); err != nil {
						return err
					}
				}

				// If we start from the app...
				if c.String("from") == "app" {
					if err := startFromApp(env); err != nil {
						return err
					}
				} else {
					if err := start(env, true); err != nil {
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
				env, err := newClientEnv(c)
				if err != nil {
					return err
				}
				if err := stop(env); err != nil {
					return err
				}
				fmt.Println("stopped")
				return nil
			},
		},
		cli.Command{
			Name:  "uninstall",
			Usage: "Uninstall",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "force", Usage: "force"},
			},
			Action: func(c *cli.Context) error {
				env, err := newClientEnv(c)
				if err != nil {
					return err
				}

				if !c.Bool("force") {
					reader := bufio.NewReader(os.Stdin)
					words := keys.RandWords(6)
					fmt.Printf("Are you sure you want to uninstall and remove your vault?\n")
					fmt.Printf("If so enter this phrase: %s\n\n", words)
					text, _ := reader.ReadString('\n')
					text = strings.Trim(text, "\r\n")
					fmt.Println("")
					if text != words {
						fmt.Println("Phrase doesn't match.")
						os.Exit(1)
					}
				}

				return Uninstall(os.Stderr, env)
			},
		},
	}
}
