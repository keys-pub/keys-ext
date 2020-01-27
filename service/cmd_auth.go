package service

import (
	"context"
	"fmt"
	"os"
	strings "strings"

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
			},
			Aliases: []string{"unlock"},
			Action: func(c *cli.Context) error {
				status, err := client.ProtoClient().RuntimeStatus(context.TODO(), &RuntimeStatusRequest{})
				if err != nil {
					return err
				}
				setupNeeded := status.AuthSetupNeeded
				logger.Infof("Auth setup needed? %t", setupNeeded)

				password := []byte(c.String("password"))

				var authToken string
				if setupNeeded {
					logger.Infof("Auth setup...")

					if len(password) == 0 {
						fmt.Fprintf(os.Stderr, "OK, let's create a password.\n")
						var err error
						password, err = readPassword("Create a password:", true)
						if err != nil {
							return err
						}
					}

					setupResp, err := client.ProtoClient().AuthSetup(context.TODO(), &AuthSetupRequest{
						Password: string(password),
					})
					if err != nil {
						return err
					}
					authToken = setupResp.AuthToken

				} else {
					if len(password) == 0 {
						var err error
						password, err = readPassword("Enter your password:", false)
						if err != nil {
							return err
						}
					}

					logger.Infof("Auth unlock...")
					unlock, unlockErr := client.ProtoClient().AuthUnlock(context.TODO(), &AuthUnlockRequest{
						Password: string(password),
						Client:   "cli",
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
				_, lockErr := client.ProtoClient().AuthLock(context.TODO(), &AuthLockRequest{})
				if lockErr != nil {
					return lockErr
				}
				return nil
			},
		},
	}
}

func wordWrap(text string, lineWidth int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}
	wrapped := words[0]
	spaceLeft := lineWidth - len(wrapped)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped += "\n" + word
			spaceLeft = lineWidth - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}
	return wrapped
}
