package mock_test

import (
	"context"
	"testing"

	"github.com/keys-pub/keysd/auth/fido2"
	"github.com/keys-pub/keysd/auth/fido2/mock"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	ctx := context.TODO()
	server := mock.NewAuthServer()

	resp, err := server.Devices(ctx, &fido2.DevicesRequest{})
	require.NoError(t, err)

	for _, device := range resp.Devices {
		require.NotEmpty(t, device.Path)

		t.Logf("Device: %s", device.Path)
		infoResp, err := server.DeviceInfo(ctx, &fido2.DeviceInfoRequest{
			Device: device.Path,
		})
		require.NoError(t, err)
		t.Logf("Info: %+v", infoResp.Info)
		require.NotEmpty(t, infoResp.Info.AAGUID)
	}
}
