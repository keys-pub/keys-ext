package rpc_test

import (
	context "context"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/keys-pub/keys-ext/auth/fido2"
	"github.com/keys-pub/keys-ext/auth/rpc"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	ctx := context.TODO()
	server := rpc.NewAuthServer()

	resp, err := server.Devices(ctx, &fido2.DevicesRequest{})
	require.NoError(t, err)

	for _, device := range resp.Devices {
		t.Logf("Device: %+v", device.Path)
		require.NotEmpty(t, device.Path)

		typeResp, err := server.DeviceType(ctx, &fido2.DeviceTypeRequest{
			Device: device.Path,
		})
		require.NoError(t, err)
		if typeResp.Type != fido2.FIDO2 {
			continue
		}

		infoResp, err := server.DeviceInfo(ctx, &fido2.DeviceInfoRequest{
			Device: device.Path,
		})
		require.NoError(t, err)
		t.Logf("Info: %+v", infoResp.Info)
		require.NotEmpty(t, infoResp.Info.AAGUID)
	}
}

func TestConcurrent(t *testing.T) {
	ctx := context.TODO()
	server := rpc.NewAuthServer()

	resp, err := server.Devices(ctx, &fido2.DevicesRequest{})
	require.NoError(t, err)

	wg := &sync.WaitGroup{}

	fn := func() {
		defer wg.Done()
		for _, device := range resp.Devices {
			t.Logf("Device: %+v", device.Path)
			require.NotEmpty(t, device.Path)

			typeResp, err := server.DeviceType(ctx, &fido2.DeviceTypeRequest{
				Device: device.Path,
			})
			require.NoError(t, err)
			if typeResp.Type != fido2.FIDO2 {
				continue
			}

			infoResp, err := server.DeviceInfo(ctx, &fido2.DeviceInfoRequest{
				Device: device.Path,
			})
			require.NoError(t, err)
			t.Logf("Info: %+v", infoResp.Info)
			require.NotEmpty(t, infoResp.Info.AAGUID)
		}
	}

	wg.Add(5)
	go fn()
	go fn()
	go fn()
	go fn()
	go fn()

	wg.Wait()
}

func ExampleAuthServer_SetPIN() {
	if os.Getenv("FIDO2_EXAMPLES") != "1" {
		return
	}

	ctx := context.TODO()
	server := rpc.NewAuthServer()

	resp, err := server.Devices(ctx, &fido2.DevicesRequest{})
	if err != nil {
		log.Fatal(err)
	}
	if len(resp.Devices) < 1 {
		log.Fatal("No devices")
	}
	if len(resp.Devices) != 1 {
		log.Fatal("Too many devices")
	}
	path := resp.Devices[0].Path

	_, err = server.SetPIN(ctx, &fido2.SetPINRequest{
		Device: path,
		PIN:    "12345",
		OldPIN: "",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	//
}

func ExampleAuthServer_Credentials() {
	if os.Getenv("FIDO2_EXAMPLES") != "1" {
		return
	}

	ctx := context.TODO()
	server := rpc.NewAuthServer()

	dresp, err := server.Devices(ctx, &fido2.DevicesRequest{})
	if err != nil {
		log.Fatal(err)
	}
	if len(dresp.Devices) < 1 {
		log.Fatal("No devices")
	}
	if len(dresp.Devices) != 1 {
		log.Fatal("Too many devices")
	}
	path := dresp.Devices[0].Path

	cresp, err := server.Credentials(ctx, &fido2.CredentialsRequest{
		Device: path,
		PIN:    "12345",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, cred := range cresp.Credentials {
		log.Printf("%+v\n", cred)
	}

	// Output:
	//
}

func ExampleAuthServer_Reset() {
	if os.Getenv("FIDO2_EXAMPLES_RESET") != "1" {
		return
	}

	ctx := context.TODO()
	server := rpc.NewAuthServer()

	dr, err := server.Devices(ctx, &fido2.DevicesRequest{})
	if err != nil {
		log.Fatal(err)
	}
	if len(dr.Devices) < 1 {
		log.Fatal("No devices")
	}
	if len(dr.Devices) != 1 {
		log.Fatal("Too many devices")
	}
	path := dr.Devices[0].Path

	_, err = server.Reset(ctx, &fido2.ResetRequest{
		Device: path,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	//
}
