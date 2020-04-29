package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/keys-pub/keysd/fido2"
	"github.com/urfave/cli"
)

func fido2Commands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "fido2",
			Usage: "FIDO2",
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "devices",
					Usage: "Show devices",
					Flags: []cli.Flag{},
					Action: func(c *cli.Context) error {
						req := &fido2.DevicesRequest{}
						resp, err := client.FIDO2Client().Devices(context.TODO(), req)
						if err != nil {
							return err
						}
						printResponse(resp)
						return nil
					},
				},
				cli.Command{
					Name:      "device-info",
					Usage:     "Device info",
					Flags:     []cli.Flag{},
					ArgsUsage: "<device path>",
					Action: func(c *cli.Context) error {
						req := fido2.DeviceInfoRequest{
							Device: c.Args().First(),
						}
						resp, err := client.FIDO2Client().DeviceInfo(context.TODO(), &req)
						if err != nil {
							return err
						}
						printResponse(resp)
						return nil
					},
				},
			},
		},
	}
}

func printResponse(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
