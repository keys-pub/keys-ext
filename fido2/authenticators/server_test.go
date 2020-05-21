package authenticators_test

import (
	context "context"
	"log"
	"os"
	"testing"

	"github.com/keys-pub/keysd/fido2"
	"github.com/keys-pub/keysd/fido2/authenticators"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	ctx := context.TODO()
	server := authenticators.NewAuthenticatorsServer()

	resp, err := server.Devices(ctx, &fido2.DevicesRequest{})
	require.NoError(t, err)

	for _, device := range resp.Devices {
		require.NotEmpty(t, device.Path)

		infoResp, err := server.DeviceInfo(ctx, &fido2.DeviceInfoRequest{
			Device: device.Path,
		})
		require.NoError(t, err)
		t.Logf("Info: %+v", infoResp.Info)
		require.NotEmpty(t, infoResp.Info.AAGUID)
	}
}

func ExampleAuthenticatorsServer_SetPIN() {
	if os.Getenv("FIDO2_EXAMPLES") != "1" {
		return
	}

	ctx := context.TODO()
	server := authenticators.NewAuthenticatorsServer()

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
		PIN:    "123456",
		OldPIN: "sdflfwdsasdfsadf",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	//
}

func ExampleAuthenticatorsServer_Credentials() {
	if os.Getenv("FIDO2_EXAMPLES") != "1" {
		return
	}

	ctx := context.TODO()
	server := authenticators.NewAuthenticatorsServer()

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
