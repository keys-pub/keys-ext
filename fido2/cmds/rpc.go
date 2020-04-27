package cmds

import (
	"context"
	"encoding/json"

	"github.com/keys-pub/keysd/fido2"
	"github.com/urfave/cli"
)

type devicesFn func(ctx context.Context, in *fido2.DevicesRequest) (*fido2.DevicesResponse, error)

func devices(rpc devicesFn) cli.Command {
	return cli.Command{
		Name:  "devices",
		Usage: "Show devices",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			req := &fido2.DevicesRequest{}
			resp, err := rpc(context.TODO(), req)
			if err != nil {
				return err
			}
			printResponse(resp)
			return nil
		},
	}
}

// Devices ...
func Devices(rpc fido2.AuthenticatorsServer) cli.Command {
	return devices(rpc.Devices)
}

// DevicesFn ...
func DevicesFn(rpc func() fido2.AuthenticatorsClient) cli.Command {
	return devices(func(ctx context.Context, in *fido2.DevicesRequest) (*fido2.DevicesResponse, error) {
		return rpc().Devices(ctx, in)
	})
}

type deviceInfoFn func(ctx context.Context, in *fido2.DeviceInfoRequest) (*fido2.DeviceInfoResponse, error)

func deviceInfo(rpc deviceInfoFn) cli.Command {
	return cli.Command{
		Name:      "device-info",
		Usage:     "Device info",
		Flags:     []cli.Flag{},
		ArgsUsage: deviceInfoRequestExample(),
		Action: func(c *cli.Context) error {
			var req fido2.DeviceInfoRequest
			if err := json.Unmarshal([]byte(c.Args().First()), &req); err != nil {
				return err
			}
			resp, err := rpc(context.TODO(), &req)
			if err != nil {
				return err
			}
			printResponse(resp)
			return nil
		},
	}
}

// DeviceInfo ...
func DeviceInfo(rpc fido2.AuthenticatorsServer) cli.Command {
	return deviceInfo(rpc.DeviceInfo)
}

// DeviceInfoFn ...
func DeviceInfoFn(rpc func() fido2.AuthenticatorsClient) cli.Command {
	return deviceInfo(func(ctx context.Context, in *fido2.DeviceInfoRequest) (*fido2.DeviceInfoResponse, error) {
		return rpc().DeviceInfo(ctx, in)
	})
}

func deviceInfoRequestExample() string {
	req := fido2.DeviceInfoRequest{
		Device: "device path or name",
	}
	b, err := json.Marshal(req) // , "   ", "  ")
	if err != nil {
		panic(err)
	}
	return "'" + string(b) + "'"
}

type makeCredentialFn func(ctx context.Context, in *fido2.MakeCredentialRequest) (*fido2.MakeCredentialResponse, error)

func makeCredential(rpc makeCredentialFn) cli.Command {
	return cli.Command{
		Name:      "make-credential",
		Usage:     "Make credential",
		Flags:     []cli.Flag{},
		ArgsUsage: makeCredentialRequestExample(),
		Action: func(c *cli.Context) error {
			var req fido2.MakeCredentialRequest
			if err := json.Unmarshal([]byte(c.Args().First()), &req); err != nil {
				return err
			}
			resp, err := rpc(context.TODO(), &req)
			if err != nil {
				return err
			}
			printResponse(resp)
			return nil
		},
	}
}

// MakeCredential ...
func MakeCredential(rpc fido2.AuthenticatorsServer) cli.Command {
	return makeCredential(rpc.MakeCredential)
}

// MakeCredentialFn ...
func MakeCredentialFn(rpc func() fido2.AuthenticatorsClient) cli.Command {
	return makeCredential(func(ctx context.Context, in *fido2.MakeCredentialRequest) (*fido2.MakeCredentialResponse, error) {
		return rpc().MakeCredential(ctx, in)
	})
}

func makeCredentialRequestExample() string {
	req := fido2.MakeCredentialRequest{
		Device:         "<device path or name>",
		ClientDataHash: []byte{0x01},
		RP: &fido2.RelyingParty{
			ID:   "rpID",
			Name: "rpName",
		},
		User: &fido2.User{
			ID:          []byte{0x02},
			Name:        "userName",
			DisplayName: "userDisplayName",
		},
		Type: "es256", // Algorithm
		PIN:  "12345", // Pin
	}
	b, err := json.Marshal(req) // , "   ", "  ")
	if err != nil {
		panic(err)
	}
	return "'" + string(b) + "'"
}
