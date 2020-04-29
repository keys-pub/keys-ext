package libfido2_test

import (
	context "context"
	"log"
	"os"
	"testing"

	"github.com/keys-pub/keysd/fido2"
	"github.com/keys-pub/keysd/fido2/libfido2"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	ctx := context.TODO()
	server := libfido2.NewAuthenticatorsServer()

	resp, err := server.Devices(ctx, &fido2.DevicesRequest{})
	require.NoError(t, err)

	for _, device := range resp.Devices {
		require.NotEmpty(t, device.Path)

		infoResp, err := server.DeviceInfo(ctx, &fido2.DeviceInfoRequest{
			Device: device.Path,
		})
		require.NoError(t, err)
		t.Logf("Info: %+v", infoResp.Info)
	}
}

func ExampleAuthenticatorsServer_SetPIN() {
	if os.Getenv("FIDO2_EXAMPLES") != "1" {
		return
	}

	ctx := context.TODO()
	server := libfido2.NewAuthenticatorsServer()

	resp, err := server.Devices(ctx, &fido2.DevicesRequest{})
	if err != nil {
		log.Fatal(err)
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
}
