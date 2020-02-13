package service

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

func importCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:      "import",
			Usage:     "Import a key",
			ArgsUsage: "<stdin or -in>",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "in, i", Usage: "file to read or stdin if not specified"},
				cli.StringFlag{Name: "password, p", Usage: "password"},
			},
			Action: func(c *cli.Context) error {
				inPath := c.String("in")
				var b []byte
				if inPath != "" {
					inPath, err := filepath.Abs(inPath)
					if err != nil {
						return err
					}
					in, err := ioutil.ReadFile(inPath)
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

				password := c.String("password")
				if len(password) == 0 {
					p, err := readPassword("Enter the password:")
					if err != nil {
						return err
					}
					password = p
				}

				req := &KeyImportRequest{
					In:       b,
					Password: password,
				}
				resp, err := client.ProtoClient().KeyImport(context.TODO(), req)
				if err != nil {
					return err
				}
				fmt.Println(resp.KID)
				return nil
			},
		},
	}
}
