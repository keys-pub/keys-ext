package fido2

import (
	"context"
	"sort"

	"github.com/keys-pub/go-libfido2"
)

type service struct{}

// NewAuthenticatorsServer creates an AuthenticatorsServer.
func NewAuthenticatorsServer() AuthenticatorsServer {
	return &service{}
}

func (s *service) Devices(ctx context.Context, req *DevicesRequest) (*DevicesResponse, error) {
	devices, err := libfido2.DeviceLocations()
	if err != nil {
		return nil, err
	}
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Product < devices[j].Product
	})
	return &DevicesResponse{
		Devices: devicesToRPC(devices),
	}, nil
}

// findDevice returns a device from a name.
// You need to call Device.Close() when done.
func findDevice(name string) (*libfido2.Device, error) {
	device, err := libfido2.NewDevice(name)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (s *service) DeviceInfo(ctx context.Context, req *DeviceInfoRequest) (*DeviceInfoResponse, error) {
	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}
	defer device.Close()

	info, err := device.Info()
	if err != nil {
		return nil, err
	}

	return &DeviceInfoResponse{
		Info: deviceInfoToRPC(info),
	}, nil
}

func (s *service) MakeCredential(ctx context.Context, req *MakeCredentialRequest) (*MakeCredentialResponse, error) {
	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}
	defer device.Close()

	typ, err := credTypeFromRPC(req.Type)
	if err != nil {
		return nil, err
	}
	extensions, err := extensionsFromRPC(req.Extensions)
	if err != nil {
		return nil, err
	}
	rk, err := optionValueFromRPC(req.RK)
	if err != nil {
		return nil, err
	}
	uv, err := optionValueFromRPC(req.UV)
	if err != nil {
		return nil, err
	}

	attestation, err := device.MakeCredential(
		req.ClientDataHash,
		rpFromRPC(req.RP),
		userFromRPC(req.User),
		typ,
		req.PIN,
		&libfido2.MakeCredentialOpts{
			Extensions: extensions,
			RK:         rk,
			UV:         uv,
		},
	)
	if err != nil {
		return nil, err
	}
	return &MakeCredentialResponse{
		Attestation: attestationToRPC(attestation),
	}, nil
}

func (s *service) SetPIN(ctx context.Context, req *SetPINRequest) (*SetPINResponse, error) {
	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}
	defer device.Close()

	if err := device.SetPIN(req.PIN, req.OldPIN); err != nil {
		return nil, err
	}

	return &SetPINResponse{}, nil
}

func (s *service) Reset(ctx context.Context, req *ResetRequest) (*ResetResponse, error) {
	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}
	defer device.Close()

	if err := device.Reset(); err != nil {
		return nil, err
	}

	return &ResetResponse{}, nil
}

func (s *service) RetryCount(ctx context.Context, req *RetryCountRequest) (*RetryCountResponse, error) {
	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}
	defer device.Close()

	count, err := device.RetryCount()
	if err != nil {
		return nil, err
	}

	return &RetryCountResponse{
		Count: int32(count),
	}, nil
}

func (s *service) Assertion(ctx context.Context, req *AssertionRequest) (*AssertionResponse, error) {
	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}
	defer device.Close()

	extensions, err := extensionsFromRPC(req.Extensions)
	if err != nil {
		return nil, err
	}
	uv, err := optionValueFromRPC(req.UV)
	if err != nil {
		return nil, err
	}
	up, err := optionValueFromRPC(req.UP)
	if err != nil {
		return nil, err
	}

	assertion, err := device.Assertion(req.RPID, req.ClientDataHash, req.CredID, req.PIN, &libfido2.AssertionOpts{Extensions: extensions, UV: uv, UP: up})
	if err != nil {
		return nil, err
	}

	return &AssertionResponse{
		Assertion: assertionToRPC(assertion),
	}, nil
}

func (s *service) CredentialsInfo(ctx context.Context, req *CredentialsInfoRequest) (*CredentialsInfoResponse, error) {
	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}
	defer device.Close()

	info, err := device.CredentialsInfo(req.PIN)
	if err != nil {
		return nil, err
	}

	return &CredentialsInfoResponse{
		Info: credentialsInfoToRPC(info),
	}, nil
}

func (s *service) Credentials(ctx context.Context, req *CredentialsRequest) (*CredentialsResponse, error) {
	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}
	defer device.Close()

	credentials, err := device.Credentials(req.RPID, req.PIN)
	if err != nil {
		return nil, err
	}

	return &CredentialsResponse{
		Credentials: credentialsToRPC(credentials),
	}, nil
}

func (s *service) RelyingParties(ctx context.Context, req *RelyingPartiesRequest) (*RelyingPartiesResponse, error) {
	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}
	defer device.Close()

	rps, err := device.RelyingParties(req.PIN)
	if err != nil {
		return nil, err
	}

	return &RelyingPartiesResponse{
		Parties: relyingPartiesToRPC(rps),
	}, nil
}
