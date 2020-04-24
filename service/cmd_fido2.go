package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/tabwriter"

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
					Name:      "devices",
					Usage:     "List devices",
					Flags:     []cli.Flag{},
					ArgsUsage: "",
					Action: func(c *cli.Context) error {
						req := &fido2.DetectDevicesRequest{}
						resp, err := client.FIDO2Client().DetectDevices(context.TODO(), req)
						if err != nil {
							return err
						}
						fmtDevices(resp.Devices)
						return nil
					},
				},
			},
		},
	}
}

func fmtDevices(devices []*fido2.DeviceInfo) {
	out := &bytes.Buffer{}
	w := new(tabwriter.Writer)
	w.Init(out, 0, 8, 1, ' ', 0)
	for _, device := range devices {
		fmtDevice(w, device)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	fmt.Print(out.String())
}

func fmtDevice(w io.Writer, device *fido2.DeviceInfo) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n", device.Product, device.Manufacturer, device.Path, device.ProductID, device.VendorID)
}
