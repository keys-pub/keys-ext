package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

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
			},
			Action: func(c *cli.Context) error {
				kid, err := argString(c, "kid", false)
				if err != nil {
					return err
				}
				resp, err := client.ProtoClient().Key(context.TODO(), &KeyRequest{
					KID:  kid,
					User: c.String("user"),
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
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				req := &KeyGenerateRequest{}
				resp, err := client.ProtoClient().KeyGenerate(context.TODO(), req)
				if err != nil {
					return err
				}
				fmt.Println(resp.KID)
				return nil
			},
		},
		cli.Command{
			Name:  "export",
			Usage: "Export a key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid"},
				cli.StringFlag{Name: "password, p", Usage: "password"},
			},
			Action: func(c *cli.Context) error {
				kid, err := argString(c, "kid", false)
				if err != nil {
					return err
				}
				req := &KeyExportRequest{
					KID:      kid,
					Password: c.String("password"),
					Type:     SaltpackPwExportType,
				}
				resp, err := client.ProtoClient().KeyExport(context.TODO(), req)
				if err != nil {
					return err
				}
				fmt.Println(resp.Export)
				return nil
			},
		},
		cli.Command{
			Name:  "import",
			Usage: "Import a key",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				in, err := ioutil.ReadAll(bufio.NewReader(os.Stdin))
				if err != nil {
					return err
				}
				req := &KeyImportRequest{
					In: in,
				}
				resp, err := client.ProtoClient().KeyImport(context.TODO(), req)
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
