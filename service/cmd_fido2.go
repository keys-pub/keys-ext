package service

import (
	"github.com/keys-pub/keysd/fido2/cmds"
	"github.com/urfave/cli"
)

func fido2Commands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "fido2",
			Usage: "FIDO2",
			Subcommands: []cli.Command{
				cmds.DevicesFn(client.FIDO2Client),
				cmds.DeviceInfoFn(client.FIDO2Client),
				cmds.MakeCredentialFn(client.FIDO2Client),
			},
		},
	}
}
