package fido2

import (
	"context"
	"log"

	libfido2 "github.com/keys-pub/go-libfido2"
)

type service struct{}

// NewAuthenticatorsServer creates an AuthenticatorsServer.
func NewAuthenticatorsServer() AuthenticatorsServer {
	return &service{}
}

func (s *service) DetectDevices(ctx context.Context, req *DetectDevicesRequest) (*DetectDevicesResponse, error) {
	detected, err := libfido2.DetectDevices(100)
	if err != nil {
		log.Fatal(err)
	}
	return &DetectDevicesResponse{
		Devices: deviceInfosToRPC(detected),
	}, nil
}

func deviceInfosToRPC(infos []*libfido2.DeviceInfo) []*DeviceInfo {
	out := make([]*DeviceInfo, 0, len(infos))
	for _, d := range infos {
		out = append(out, &DeviceInfo{
			Path:         d.Path,
			ProductID:    int32(d.ProductID),
			VendorID:     int32(d.VendorID),
			Manufacturer: d.Manufacturer,
			Product:      d.Product,
		})
	}
	return out
}
