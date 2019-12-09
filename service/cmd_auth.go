package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	strings "strings"

	"github.com/urfave/cli"
)

type authMode string

const (
	autModeUnknown  authMode = ""
	authModeSetup   authMode = "AUTH_SETUP"
	authModeRecover authMode = "AUTH_RECOVER"
	authModeUnlock  authMode = "AUTH_UNLOCK"
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

				reader := bufio.NewReader(os.Stdin)
				authMode := autModeUnknown
				if setupNeeded {
					fmt.Fprintf(os.Stderr, "Would you like to setup a new key or use an existing one?\n")
					fmt.Fprintf(os.Stderr, "(n) New key\n")
					fmt.Fprintf(os.Stderr, "(e) Existing key\n")
					in, err := reader.ReadString('\n')
					if err != nil {
						return err
					}
					switch strings.TrimSpace(strings.ToLower(in)) {
					case "n", "new":
						authMode = authModeSetup
					case "e", "existing":
						authMode = authModeRecover
					}
					fmt.Fprintf(os.Stderr, "\n")
				} else {
					authMode = authModeUnlock
				}

				var authToken string
				switch authMode {
				case authModeSetup:
					logger.Infof("Auth setup...")

					if len(password) == 0 {
						fmt.Fprintf(os.Stderr, "OK, let's create a password.\n")
						var err error
						password, err = readPassword("Create a password:", true)
						if err != nil {
							return err
						}
					}

					rand, randErr := client.ProtoClient().Rand(context.TODO(), &RandRequest{
						Length:   32,
						Encoding: BIP39,
					})
					if randErr != nil {
						return randErr
					}
					pepper := string(rand.Data)

					fmt.Fprintf(os.Stderr, "\n")
					fmt.Fprintf(os.Stderr, wordWrap("Now you'll need to backup a (secret) recovery phrase. This phrase by itself can't be used by itself to access your keys. A good way to backup this phrase is to email it to yourself or save it in the cloud in a place only you can access. This allows you to recover your account if your devices go missing.", 80))
					fmt.Fprintf(os.Stderr, "\n\n")
					fmt.Fprintf(os.Stderr, "Your recovery phrase is:\n\n%s\n\n", wordWrap(pepper, 80))

				confirmRecovery:
					for {
						fmt.Fprintf(os.Stderr, wordWrap("Have you backed up this recovery phrase (y/n)?", 80)+" ")
						in, err := reader.ReadString('\n')
						if err != nil {
							return err
						}
						switch strings.TrimSpace(strings.ToLower(in)) {
						case "y", "yes":
							// TODO: Ask for phrase to double check?
							break confirmRecovery
						}
					}

					publish := false
				confirmPublish:
					for {
						fmt.Fprintf(os.Stderr, wordWrap("Do you want to publish your public key to the key server (keys.pub) (Y/n)?", 80)+" ")
						in, err := reader.ReadString('\n')
						if err != nil {
							return err
						}
						switch strings.TrimSpace(strings.ToLower(in)) {
						case "y", "yes", "":
							publish = true
							break confirmPublish
						case "n", "no":
							break confirmPublish
						default:
						}
					}

					fmt.Fprintf(os.Stderr, "\nSaving...")

					auth, err := client.ProtoClient().AuthSetup(context.TODO(), &AuthSetupRequest{
						Pepper:           pepper,
						Password:         string(password),
						PublishPublicKey: publish,
						Force:            c.Bool("force"),
						Client:           "cli",
					})
					if err != nil {
						return err
					}
					authToken = auth.AuthToken
					fmt.Fprintf(os.Stderr, "\nSaved key %s\n\n", auth.KID)

				case authModeRecover:
					if len(password) == 0 {
						var err error
						password, err = readPassword("Enter your password:", true)
						if err != nil {
							return err
						}
					}

					fmt.Fprintf(os.Stderr, "What's your recovery phrase? ")
					in, err := reader.ReadString('\n')
					if err != nil {
						return err
					}
					pepper := strings.TrimSpace(strings.ToLower(in))

					logger.Infof("Auth recover...")
					fmt.Fprintf(os.Stderr, "\nRecovering...")
					auth, err := client.ProtoClient().AuthSetup(context.TODO(), &AuthSetupRequest{
						Pepper:   pepper,
						Password: string(password),
						Recover:  true,
						Force:    c.Bool("force"),
						Client:   "cli",
					})
					if err != nil {
						return err
					}
					authToken = auth.AuthToken
					fmt.Fprintf(os.Stderr, "\nRecovered key %s\n\n", auth.KID)

				case authModeUnlock:
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
