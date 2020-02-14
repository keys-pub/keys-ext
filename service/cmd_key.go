package service

import (
	"context"
	"encoding/json"
	"fmt"
	strings "strings"

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
				resp, err := client.ProtoClient().Keys(context.TODO(), &KeysRequest{Types: types})
				if err != nil {
					return err
				}
				fmtKeys(resp.Keys)
				return nil
			},
		},
		cli.Command{
			Name:  "keyring",
			Usage: "Keyring",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				resp, err := client.ProtoClient().Items(context.TODO(), &ItemsRequest{})
				if err != nil {
					return err
				}
				fmtItems(resp.Items)
				return nil
			},
		},
		cli.Command{
			Name:      "key",
			Usage:     "Show key",
			ArgsUsage: "kid or user",
			Action: func(c *cli.Context) error {
				identity := c.Args().First()
				if identity == "" {
					return errors.Errorf("specify kid or user")
				}
				resp, err := client.ProtoClient().Key(context.TODO(), &KeyRequest{
					Identity: identity,
				})
				if err != nil {
					return err
				}
				if resp.Key == nil {
					return errors.Errorf("key not found")
				}
				// fmtKeys([]*Key{resp.Key})
				// TODO: key type outputs as int
				b, err := json.MarshalIndent(resp.Key, "", "  ")
				if err != nil {
					return err
				}
				fmt.Print(string(b))
				return nil
			},
		},
		cli.Command{
			Name:  "generate",
			Usage: "Generate a key",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				// TODO: Key type
				req := &KeyGenerateRequest{
					Type: EdX25519,
				}
				resp, err := client.ProtoClient().KeyGenerate(context.TODO(), req)
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
				if _, err := client.ProtoClient().KeyRemove(context.TODO(), &KeyRemoveRequest{
					KID: kid,
				}); err != nil {
					return err
				}
				return nil
			},
		},
	}
}
