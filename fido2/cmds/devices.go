package cmds

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/keys-pub/keysd/fido2"
	"github.com/urfave/cli"
)

// Devices ...
func Devices(client fido2.AuthenticatorsClient) cli.Command {
	return cli.Command{
		Name:      "devices",
		Usage:     "Show devices",
		Flags:     []cli.Flag{},
		ArgsUsage: "",
		Action: func(c *cli.Context) error {
			req := &fido2.DeviceLocationsRequest{}
			resp, err := client.DeviceLocations(context.TODO(), req)
			if err != nil {
				return err
			}
			fmtDeviceLocations(resp.Locations)
			return nil
		},
	}
}

func fmtDeviceLocations(locs []*fido2.DeviceLocation) {
	out := &bytes.Buffer{}
	w := new(tabwriter.Writer)
	w.Init(out, 0, 8, 1, ' ', 0)
	for _, loc := range locs {
		fmtDeviceLocation(w, loc)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	fmt.Print(out.String())
}

func fmtDeviceLocation(w io.Writer, loc *fido2.DeviceLocation) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n", loc.Product, loc.Manufacturer, loc.Path, loc.ProductID, loc.VendorID)
}
