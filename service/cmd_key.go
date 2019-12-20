package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func itemCommands(client *Client) []cli.Command {
	return []cli.Command{
		// cli.Command{
		// 	Name:  "list",
		// 	Usage: "List keys",
		// 	Flags: []cli.Flag{},
		// 	Action: func(c *cli.Context) error {
		// 		resp, err := client.ProtoClient().Keys(context.TODO(), &KeysRequest{})
		// 		if err != nil {
		// 			return err
		// 		}
		// 		fmtKeys(resp.Keys)
		// 		return nil
		// 	},
		// },
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
			Name:  "key",
			Usage: "Show key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k"},
				cli.StringFlag{Name: "user, u"},
				cli.BoolFlag{Name: "skip-check"},
				cli.BoolFlag{Name: "update"},
			},
			Action: func(c *cli.Context) error {
				kid, err := argString(c, "kid", true)
				if err != nil {
					return err
				}
				resp, err := client.ProtoClient().Key(context.TODO(), &KeyRequest{
					KID:       kid,
					User:      c.String("user"),
					SkipCheck: c.Bool("skip-check"),
					Update:    c.Bool("update"),
				})
				if err != nil {
					return err
				}
				if resp.Key == nil {
					return errors.Errorf("key not found")
				}
				// fmtKeys([]*Key{resp.Key})
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
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "publish", Usage: "publish public key to key server"},
			},
			Action: func(c *cli.Context) error {
				req := &KeyGenerateRequest{
					PublishPublicKey: c.Bool("publish"),
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
			Name:  "backup",
			Usage: "Backup a key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid"},
			},
			Action: func(c *cli.Context) error {
				kid, err := argString(c, "kid", false)
				if err != nil {
					return err
				}
				req := &KeyBackupRequest{
					KID: kid,
				}
				resp, err := client.ProtoClient().KeyBackup(context.TODO(), req)
				if err != nil {
					return err
				}
				fmt.Println(resp.SeedPhrase)
				return nil
			},
		},
		cli.Command{
			Name:    "recover",
			Usage:   "Recover a key",
			Aliases: []string{"import"},
			Flags: []cli.Flag{
				cli.StringFlag{Name: "seed-phrase", Usage: "seed phrase"},
				cli.BoolFlag{Name: "publish", Usage: "publish public key to key server"},
			},
			Action: func(c *cli.Context) error {
				seedPhrase := c.String("seed-phrase")
				req := &KeyRecoverRequest{
					SeedPhrase:       seedPhrase,
					PublishPublicKey: c.Bool("publish"),
				}
				resp, err := client.ProtoClient().KeyRecover(context.TODO(), req)
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
				cli.StringFlag{Name: "seed-phrase", Usage: "seed phrase for confirming removal of a key"},
			},
			Action: func(c *cli.Context) error {
				kid, err := argString(c, "kid", false)
				if err != nil {
					return err
				}
				if _, err := client.ProtoClient().KeyRemove(context.TODO(), &KeyRemoveRequest{
					KID:        kid,
					SeedPhrase: c.String("seed-phrase"),
				}); err != nil {
					return err
				}
				return nil
			},
		},
		cli.Command{
			Name:  "share",
			Usage: "Share a key with a recipient",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid"},
				cli.StringFlag{Name: "recipient, r", Usage: "recipient"},
			},
			Action: func(c *cli.Context) error {
				if c.String("kid") == "" {
					return errors.Errorf("no kid specified")
				}
				if c.String("recipient") == "" {
					return errors.Errorf("no recipient specified")
				}
				_, err := client.ProtoClient().KeyShare(context.TODO(), &KeyShareRequest{
					KID:       c.String("kid"),
					Recipient: c.String("recipient"),
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		cli.Command{
			Name:  "retrieve",
			Usage: "Retrieve a (shared) key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid"},
				cli.StringFlag{Name: "recipient, r", Usage: "recipient"},
			},
			Action: func(c *cli.Context) error {
				if c.String("kid") == "" {
					return errors.Errorf("no kid specified")
				}
				if c.String("recipient") == "" {
					return errors.Errorf("no recipient specified")
				}
				_, err := client.ProtoClient().KeyRetrieve(context.TODO(), &KeyRetrieveRequest{
					KID:       c.String("kid"),
					Recipient: c.String("recipient"),
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
	}
}
