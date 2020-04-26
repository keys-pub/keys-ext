package fido2

import (
	"context"
	"log"

	"github.com/keys-pub/go-libfido2"
	"github.com/pkg/errors"
)

type service struct{}

// NewAuthenticatorsServer creates an AuthenticatorsServer.
func NewAuthenticatorsServer() AuthenticatorsServer {
	return &service{}
}

func (s *service) DeviceLocations(ctx context.Context, req *DeviceLocationsRequest) (*DeviceLocationsResponse, error) {
	detected, err := libfido2.DeviceLocations()
	if err != nil {
		log.Fatal(err)
	}
	return &DeviceLocationsResponse{
		Locations: deviceLocationsToRPC(detected),
	}, nil
}

func (s *service) DeviceInfo(ctx context.Context, req *DeviceInfoRequest) (*DeviceInfoResponse, error) {
	return &DeviceInfoResponse{}, errors.Errorf("not implemented")
}

func (s *service) MakeCredential(ctx context.Context, req *MakeCredentialRequest) (*MakeCredentialResponse, error) {
	return &MakeCredentialResponse{}, errors.Errorf("not implemented")
}

func (s *service) SetPIN(ctx context.Context, req *SetPINRequest) (*SetPINResponse, error) {
	return &SetPINResponse{}, errors.Errorf("not implemented")
}

func (s *service) Reset(ctx context.Context, req *ResetRequest) (*ResetResponse, error) {
	return &ResetResponse{}, errors.Errorf("not implemented")
}

func (s *service) RetryCount(ctx context.Context, req *RetryCountRequest) (*RetryCountResponse, error) {
	return &RetryCountResponse{}, errors.Errorf("not implemented")
}

func (s *service) Assertion(ctx context.Context, req *AssertionRequest) (*AssertionResponse, error) {
	return &AssertionResponse{}, errors.Errorf("not implemented")
}

func (s *service) CredentialsInfo(ctx context.Context, req *CredentialsInfoRequest) (*CredentialsInfoResponse, error) {
	return &CredentialsInfoResponse{}, errors.Errorf("not implemented")
}

func (s *service) Credentials(ctx context.Context, req *CredentialsRequest) (*CredentialsResponse, error) {
	return &CredentialsResponse{}, errors.Errorf("not implemented")
}

func (s *service) RelyingParties(ctx context.Context, req *RelyingPartiesRequest) (*RelyingPartiesResponse, error) {
	return &RelyingPartiesResponse{}, errors.Errorf("not implemented")
}

func deviceLocationsToRPC(locs []*libfido2.DeviceLocation) []*DeviceLocation {
	out := make([]*DeviceLocation, 0, len(locs))
	for _, d := range locs {
		out = append(out, &DeviceLocation{
			Path:         d.Path,
			ProductID:    int32(d.ProductID),
			VendorID:     int32(d.VendorID),
			Manufacturer: d.Manufacturer,
			Product:      d.Product,
		})
	}
	return out
}
