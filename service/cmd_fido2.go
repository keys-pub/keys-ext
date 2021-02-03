package service

import (
	"context"

	"github.com/keys-pub/keys-ext/auth/fido2"
	"github.com/urfave/cli"
)

func fido2Commands(client *Client) []cli.Command {
	return []cli.Command{
		{
			Name:  "fido2",
			Usage: "FIDO2",
			Subcommands: []cli.Command{
				{
					Name:  "devices",
					Usage: "Show devices",
					Flags: []cli.Flag{},
					Action: func(c *cli.Context) error {
						req := &fido2.DevicesRequest{}
						resp, err := client.FIDO2Client().Devices(context.TODO(), req)
						if err != nil {
							return err
						}
						printMessage(resp)
						return nil
					},
				},
				{
					Name:  "device-info",
					Usage: "Device info",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "device, d", Usage: "device path"},
					},
					Action: func(c *cli.Context) error {
						req := fido2.DeviceInfoRequest{
							Device: c.String("device"),
						}
						resp, err := client.FIDO2Client().DeviceInfo(context.TODO(), &req)
						if err != nil {
							return err
						}
						printMessage(resp)
						return nil
					},
				},
				{
					Name:  "credentials-info",
					Usage: "Credentials info",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "device, d", Usage: "device path"},
						cli.StringFlag{Name: "pin", Usage: "pin"},
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
						printMessage(resp)
						return nil
					},
				},
				{
					Name:  "credentials",
					Usage: "Credentials",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "device, d", Usage: "device"},
						cli.StringFlag{Name: "pin", Usage: "pin"},
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
						printMessage(resp)
						return nil
					},
				},
			},
		},
	}
}
