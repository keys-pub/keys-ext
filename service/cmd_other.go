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
			Usage: "Random bytes",
			Flags: []cli.Flag{
				cli.IntFlag{Name: "n", Usage: "number of bytes", Value: 32},
				cli.StringFlag{Name: "encoding, enc, e", Usage: "encoding (base64, base62, base58, base32, base16, bip39, saltpack)", Value: "base62"},
			},
			Action: func(c *cli.Context) error {
				enc, err := encodingToRPC(c.String("enc"))
				if err != nil {
					return err
				}
				rand, err := client.ProtoClient().Rand(context.TODO(), &RandRequest{
					NumBytes: int32(c.Int("n")),
					Encoding: enc,
				})
				if err != nil {
					return err
				}
				fmt.Println(string(rand.Data))
				return nil
			},
		},
	}
}
