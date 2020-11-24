package service

import (
	"context"
	"fmt"

	"github.com/urfave/cli"
)

func otherCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "rand",
			Usage: "Generate randomness",
			Flags: []cli.Flag{
				cli.IntFlag{Name: "num-bytes, n", Usage: "number of bytes", Value: 32},
				cli.StringFlag{Name: "encoding, enc, e", Usage: "encoding (base64, base62, base58, base32, base16, hex, bip39, saltpack)", Value: "base62"},
				cli.BoolFlag{Name: "no-padding", Usage: "no padding (base64, base32"},
				cli.BoolFlag{Name: "lower", Usage: "lowercase (base64, base32"},
			},
			Action: func(c *cli.Context) error {
				enc, err := encodingToRPC(c.String("enc"))
				if err != nil {
					return err
				}
				rand, err := client.KeysClient().Rand(context.TODO(), &RandRequest{
					NumBytes:  int32(c.Int("num-bytes")),
					Encoding:  enc,
					NoPadding: c.Bool("no-padding"),
					Lowercase: c.Bool("lower"),
				})
				if err != nil {
					return err
				}
				fmt.Println(string(rand.Data))
				return nil
			},
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "password",
					Usage: "Generate random password",
					Flags: []cli.Flag{
						cli.IntFlag{Name: "length, l", Usage: "length of password", Value: 16},
					},
					Action: func(c *cli.Context) error {
						rand, err := client.KeysClient().RandPassword(context.TODO(), &RandPasswordRequest{
							Length: int32(c.Int("length")),
						})
						if err != nil {
							return err
						}
						fmt.Println(rand.Password)
						return nil
					},
				},
				cli.Command{
					Name:  "passphrase",
					Usage: "Generate random passphrase",
					Flags: []cli.Flag{},
					Action: func(c *cli.Context) error {
						rand, err := client.KeysClient().Rand(context.TODO(), &RandRequest{
							NumBytes: 16,
							Encoding: BIP39,
						})
						if err != nil {
							return err
						}
						fmt.Println(string(rand.Data))
						return nil
					},
				},
			},
		},
	}
}
