package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func keyCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "list",
			Usage: "List keys",
			Flags: []cli.Flag{
				cli.StringSliceFlag{Name: "type, t", Usage: "only these types (" + strings.Join(keyTypeStrings, ", ") + ")"},
			},
			Action: func(c *cli.Context) error {
				types := []KeyType{}
				for _, t := range c.StringSlice("type") {
					typ, err := parseKeyType(t)
					if err != nil {
						return err
					}
					types = append(types, typ)
				}
				resp, err := client.KeysClient().Keys(context.TODO(), &KeysRequest{Types: types})
				if err != nil {
					return err
				}
				fmtKeys(os.Stdout, resp.Keys)
				return nil
			},
		},
		cli.Command{
			Name:      "key",
			Usage:     "Show key",
			ArgsUsage: "kid or user",
			Action: func(c *cli.Context) error {
				key := c.Args().First()
				if key == "" {
					return errors.Errorf("specify kid or user@service")
				}
				resp, err := client.KeysClient().Key(context.TODO(), &KeyRequest{
					Key: key,
				})
				if err != nil {
					return err
				}
				if resp.Key == nil {
					return errors.Errorf("key not found")
				}
				printMessage(resp.Key)
				return nil
			},
		},
		cli.Command{
			Name:  "generate",
			Usage: "Generate a key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "type, t", Usage: "type (edx25519, x25519)"},
			},
			Action: func(c *cli.Context) error {
				var typ KeyType
				switch c.String("type") {
				case "", "edx25519":
					typ = EdX25519
				case "x25519":
					typ = X25519
				default:
					return errors.Errorf("unrecognized key type")
				}

				req := &KeyGenerateRequest{
					Type: typ,
				}
				resp, err := client.KeysClient().KeyGenerate(context.TODO(), req)
				if err != nil {
					return err
				}
				fmt.Println(resp.KID)
				return nil
			},
		},
		cli.Command{
			Name:  "remove",
			Usage: "Remove a key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid"},
			},
			Action: func(c *cli.Context) error {
				kid, err := argString(c, "kid", false)
				if err != nil {
					return err
				}
				if _, err := client.KeysClient().KeyRemove(context.TODO(), &KeyRemoveRequest{
					KID: kid,
				}); err != nil {
					return err
				}
				return nil
			},
		},
	}
}
