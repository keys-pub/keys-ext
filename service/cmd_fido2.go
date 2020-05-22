package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/keys-pub/keysd/auth/fido2"
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
					Name:  "device-info",
					Usage: "Device info",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "device, d", Usage: "Device"},
					},
					Action: func(c *cli.Context) error {
						req := fido2.DeviceInfoRequest{
							Device: c.String("device"),
						}
						resp, err := client.FIDO2Client().DeviceInfo(context.TODO(), &req)
						if err != nil {
							return err
						}
						printResponse(resp)
						return nil
					},
				},
				cli.Command{
					Name:  "credentials-info",
					Usage: "Credentials info",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "device, d", Usage: "Device"},
						cli.StringFlag{Name: "pin", Usage: "PIN"},
					},
					Action: func(c *cli.Context) error {
						req := fido2.CredentialsInfoRequest{
							Device: c.String("device"),
							PIN:    c.String("pin"),
						}
						resp, err := client.FIDO2Client().CredentialsInfo(context.TODO(), &req)
						if err != nil {
							return err
						}
						printResponse(resp)
						return nil
					},
				},
				cli.Command{
					Name:  "credentials",
					Usage: "Credentials",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "device, d", Usage: "Device"},
						cli.StringFlag{Name: "pin", Usage: "PIN"},
					},
					Action: func(c *cli.Context) error {
						req := fido2.CredentialsRequest{
							Device: c.String("device"),
							PIN:    c.String("pin"),
						}
						resp, err := client.FIDO2Client().Credentials(context.TODO(), &req)
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
