package service

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/keys-pub/keys"
	"github.com/urfave/cli"
)

func importCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "import",
			Usage: "Import a key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "in, i", Usage: "file to read"},
				cli.StringFlag{Name: "password, p", Usage: "password"},
			},
			Action: func(c *cli.Context) error {
				var b []byte
				if c.String("in") != "" {
					path, err := filepath.Abs(c.String("in"))
					if err != nil {
						return err
					}
					in, err := ioutil.ReadFile(path) // #nosec
					if err != nil {
						return err
					}
					b = in
				} else {
					in, err := ioutil.ReadAll(bufio.NewReader(os.Stdin))
					if err != nil {
						return err
					}
					b = in
				}

				b, typ := keys.DetectDataType(b)
				if typ == keys.IDType {
					req := &KeyImportRequest{
						In: b,
					}
					resp, err := client.KeysClient().KeyImport(context.TODO(), req)
					if err != nil {
						return err
					}
					fmt.Println(resp.KID)
					return nil
				}

				password := c.String("password")
				if len(password) == 0 {
					p, err := readPassword("Enter the password:", false)
					if err != nil {
						return err
					}
					password = p
				}

				req := &KeyImportRequest{
					In:       b,
					Password: password,
				}
				resp, err := client.KeysClient().KeyImport(context.TODO(), req)
				if err != nil {
					return err
				}
				fmt.Println(resp.KID)
				return nil
			},
		},
	}
}
