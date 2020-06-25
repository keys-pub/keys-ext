package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func authCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "auth",
			Usage: "Authorize",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "password, pin", Usage: "password or pin"},
				cli.BoolFlag{Name: "token", Usage: "output token only"},
				cli.StringFlag{Name: "type, t", Usage: "auth type: password, fido2-hmac-secret", Value: "password"},
				cli.StringFlag{Name: "client", Value: "cli", Hidden: true},
			},
			Aliases: []string{"unlock"},
			Subcommands: []cli.Command{
				authProvisionCommand(client),
				authProvisionsCommand(client),
				authDeprovisionCommand(client),
				authVaultCommand(client),
				changePasswordCommand(client),
			},
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
				setupNeeded := status.AuthStatus == AuthSetup
				logger.Infof("Auth setup needed: %t", setupNeeded)

				clientName := c.String("client")
				if clientName == "" {
					return errors.Errorf("no client name")
				}

				authType, err := chooseAuth("How do you want to authorize?", c.String("type"))
				if err != nil {
					return err
				}

				var authToken string
				var authErr error
				if setupNeeded {
					logger.Infof("Auth setup...")
					switch authType {
					case PasswordAuth:
						authToken, authErr = passwordAuthSetup(context.TODO(), client, clientName, c.String("password"))
					case FIDO2HMACSecretAuth:
						authToken, authErr = fido2AuthSetup(context.TODO(), client, clientName, c.String("pin"))
					}
				} else {
					logger.Infof("Auth unlock...")
					switch authType {
					case PasswordAuth:
						authToken, authErr = passwordAuthUnlock(context.TODO(), client, clientName, c.String("password"))
					case FIDO2HMACSecretAuth:
						authToken, authErr = fido2AuthUnlock(context.TODO(), client, clientName, c.String("pin"))
					}
				}
				if authErr != nil {
					return authErr
				}

				if c.Bool("token") {
					fmt.Println(authToken)
					return nil
				}

				fmt.Printf("export KEYS_AUTH=\"%s\"\n", authToken)
				fmt.Printf("# For shell:\n")
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
				_, err := client.KeysClient().AuthLock(context.TODO(), &AuthLockRequest{})
				if err != nil {
					return err
				}
				return nil
			},
		},
	}
}

func authVaultCommand(client *Client) cli.Command {
	return cli.Command{
		Name:  "vault",
		Usage: "Auth from remote vault",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "key", Usage: "Key for remote vault"},
		},
		Action: func(c *cli.Context) error {
			key, err := readPassword("Vault key:", false)
			if err != nil {
				return err
			}
			if _, err := client.KeysClient().AuthVault(context.TODO(), &AuthVaultRequest{
				Key: key,
			}); err != nil {
				return err
			}
			return nil
		},
	}
}

func changePasswordCommand(client *Client) cli.Command {
	return cli.Command{
		Name:  "change-password",
		Usage: "Change password",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			old, err := readPassword("Old password:", false)
			if err != nil {
				return err
			}
			new, err := readVerifyPassword("New password:")
			if err != nil {
				return err
			}
			if _, err := client.KeysClient().PasswordChange(context.TODO(), &PasswordChangeRequest{
				Old: old,
				New: new,
			}); err != nil {
				return err
			}
			return nil
		},
	}
}

func authProvisionCommand(client *Client) cli.Command {
	return cli.Command{
		Name:  "provision",
		Usage: "Provision",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "password, pin", Usage: "password or pin"},
			cli.StringFlag{Name: "type", Usage: "auth type: password, fido2-hmac-secret"},
			cli.StringFlag{Name: "client", Value: "cli", Hidden: true},
		},
		Action: func(c *cli.Context) error {
			rts, err := client.KeysClient().RuntimeStatus(context.TODO(), &RuntimeStatusRequest{})
			if err != nil {
				return err
			}
			switch rts.AuthStatus {
			case AuthSetup:
				return status.Error(codes.Unauthenticated, "auth setup needed")
			case AuthLocked:
				return status.Error(codes.Unauthenticated, "auth locked")
			}

			clientName := c.String("client")
			if clientName == "" {
				return errors.Errorf("no client name")
			}

			authType, err := chooseAuth("How do you want to provision?", c.String("type"))
			if err != nil {
				return err
			}

			logger.Infof("Auth provision...")
			switch authType {
			case PasswordAuth:
				if err := passwordAuthProvision(context.TODO(), client, clientName, c.String("password")); err != nil {
					return err
				}
			case FIDO2HMACSecretAuth:
				pin := c.String("pin")
				if len(pin) == 0 {
					p, err := readPassword("Enter your PIN:", true)
					if err != nil {
						return err
					}
					pin = p
				}

				// Setup
				if err := fido2AuthProvision(context.TODO(), client, clientName, pin, true); err != nil {
					return err
				}
				// Unlock
				if err := fido2AuthProvision(context.TODO(), client, clientName, pin, false); err != nil {
					return err
				}

				fmt.Println("We successfully provisioned the security key (using FIDO2 hmac-secret).")
			}

			return nil
		},
	}
}

func chooseAuth(title string, arg string) (AuthType, error) {
	if arg != "" {
		return authTypeFromString(arg)
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Fprintln(os.Stderr, title)
		fmt.Fprintln(os.Stderr, "(p) Password")
		fmt.Fprintln(os.Stderr, "(f) FIDO2 hmac-secret")
		input, err := reader.ReadString('\n')
		if err != nil {
			return UnknownAuth, err
		}

		authType, err := authTypeFromString(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			return authType, nil
		}
	}
}

func authTypeFromString(s string) (AuthType, error) {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "p", "password":
		return PasswordAuth, nil
	case "f", "fido2-hmac-secret":
		return FIDO2HMACSecretAuth, nil
	default:
		return UnknownAuth, errors.Errorf("unknown auth type: %s", s)
	}
}

func authProvisionsCommand(client *Client) cli.Command {
	return cli.Command{
		Name:  "provisions",
		Usage: "Provisions",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			ctx := context.TODO()
			resp, err := client.KeysClient().AuthProvisions(ctx, &AuthProvisionsRequest{})
			if err != nil {
				return err
			}
			printMessage(resp)
			return nil
		},
	}
}

func authDeprovisionCommand(client *Client) cli.Command {
	return cli.Command{
		Name:  "deprovision",
		Usage: "Deprovision",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			id := c.Args().First()
			if id == "" {
				return errors.Errorf("specify a provision id")
			}
			_, err := client.KeysClient().AuthDeprovision(context.TODO(), &AuthDeprovisionRequest{
				ID: id,
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
}
