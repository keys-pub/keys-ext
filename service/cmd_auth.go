package service

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func authCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "auth",
			Usage: "Authenticate",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "password", Usage: "password"},
				cli.BoolFlag{Name: "token", Usage: "output token only"},
				cli.BoolFlag{Name: "force", Usage: "force recovery"},
				cli.StringFlag{Name: "client", Value: "cli", Hidden: true},
			},
			Aliases: []string{"unlock"},
			Action: func(c *cli.Context) error {
				if !c.GlobalBool("test") {
					if err := checkForAppConflict(); err != nil {
						logger.Warningf("%s", err)
					}
				}

				status, err := client.KeysClient().RuntimeStatus(context.TODO(), &RuntimeStatusRequest{})
				if err != nil {
					return err
				}
				setupNeeded := status.AuthSetupNeeded
				logger.Infof("Auth setup needed? %t", setupNeeded)

				password := c.String("password")
				clientName := c.String("client")
				if clientName == "" {
					return errors.Errorf("no client name")
				}

				var authToken string
				if setupNeeded {
					logger.Infof("Auth setup...")

					if len(password) == 0 {
						fmt.Fprintf(os.Stderr, "OK, let's create a password.\n")
						p, err := readVerifyPassword("Create a password:")
						if err != nil {
							return err
						}
						password = p
					}

					setupResp, err := client.KeysClient().AuthSetup(context.TODO(), &AuthSetupRequest{
						Password: password,
					})
					if err != nil {
						return err
					}
					authToken = setupResp.AuthToken

				} else {
					if len(password) == 0 {
						p, err := readPassword("Enter your password:")
						if err != nil {
							return err
						}
						password = p
					}

					logger.Infof("Auth unlock...")
					unlock, unlockErr := client.KeysClient().AuthUnlock(context.TODO(), &AuthUnlockRequest{
						Password: password,
						Client:   clientName,
					})
					if unlockErr != nil {
						return unlockErr
					}
					authToken = unlock.AuthToken
				}

				if c.Bool("token") {
					fmt.Println(authToken)
					return nil
				}

				fmt.Printf("export KEYS_AUTH=\"%s\"\n", authToken)
				fmt.Printf("# To include in a shell environment:\n")
				fmt.Printf("#  export KEYS_AUTH=`keys auth -token`\n")
				fmt.Printf("#\n")
				fmt.Printf("# or using eval:\n")
				fmt.Printf("#  eval $(keys auth)\n")
				fmt.Printf("#\n")
				fmt.Printf("# For Powershell:\n")
				fmt.Printf("#  $env:KEYS_AUTH = (keys auth -token)\n")

				return nil
			},
		},
		cli.Command{
			Name:  "lock",
			Usage: "Lock",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				_, lockErr := client.KeysClient().AuthLock(context.TODO(), &AuthLockRequest{})
				if lockErr != nil {
					return lockErr
				}
				return nil
			},
		},
	}
}
