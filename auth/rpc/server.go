package rpc

import (
	"context"
	"sort"
	"sync"

	"github.com/keys-pub/go-libfido2"
	"github.com/keys-pub/keys-ext/auth/fido2"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server ...
type Server struct {
	fido2.UnimplementedFIDO2Server
	sync.Mutex
}

// NewFIDO2Server creates an FIDO2Server.
func NewFIDO2Server() fido2.FIDO2Server {
	return &Server{}
}

// Devices ...
func (s *Server) Devices(ctx context.Context, req *fido2.DevicesRequest) (*fido2.DevicesResponse, error) {
	s.Lock()
	defer s.Unlock()

	devices, err := libfido2.DeviceLocations()
	if err != nil {
		return nil, err
	}
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Product < devices[j].Product
	})
	return &fido2.DevicesResponse{
		Devices: devicesToRPC(devices),
	}, nil
}

// findDevice returns a device from a name.
// You need to call Device.Close() when done.
func findDevice(path string) (*libfido2.Device, error) {
	device, err := libfido2.NewDevice(path)
	if err != nil {
		return nil, err
	}
	return device, nil
}

// DeviceType ...
func (s *Server) DeviceType(ctx context.Context, req *fido2.DeviceTypeRequest) (*fido2.DeviceTypeResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	typ, err := device.Type()
	if err != nil {
		return nil, err
	}

	var rtyp fido2.DeviceType
	switch typ {
	case libfido2.FIDO2:
		rtyp = fido2.FIDO2Device
	case libfido2.U2F:
		rtyp = fido2.U2FDevice
	default:
		rtyp = fido2.UnknownDevice
	}

	return &fido2.DeviceTypeResponse{
		Type: rtyp,
	}, nil
}

// DeviceInfo ...
func (s *Server) DeviceInfo(ctx context.Context, req *fido2.DeviceInfoRequest) (*fido2.DeviceInfoResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	info, err := device.Info()
	if err != nil {
		return nil, err
	}

	return &fido2.DeviceInfoResponse{
		Info: deviceInfoToRPC(info),
	}, nil
}

// MakeCredential ...
func (s *Server) MakeCredential(ctx context.Context, req *fido2.MakeCredentialRequest) (*fido2.MakeCredentialResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	typ, err := credentialTypeFromString(req.Type)
	if err != nil {
		return nil, err
	}
	extensions, err := extensionsFromStrings(req.Extensions)
	if err != nil {
		return nil, err
	}
	rk, err := optionValueFromString(req.RK)
	if err != nil {
		return nil, err
	}
	uv, err := optionValueFromString(req.UV)
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
	return &fido2.MakeCredentialResponse{
		Attestation: attestationToRPC(attestation),
	}, nil
}

// SetPIN ...
func (s *Server) SetPIN(ctx context.Context, req *fido2.SetPINRequest) (*fido2.SetPINResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	if err := device.SetPIN(req.PIN, req.OldPIN); err != nil {
		return nil, err
	}

	return &fido2.SetPINResponse{}, nil
}

// Reset ...
func (s *Server) Reset(ctx context.Context, req *fido2.ResetRequest) (*fido2.ResetResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	if err := device.Reset(); err != nil {
		return nil, err
	}

	return &fido2.ResetResponse{}, nil
}

// RetryCount ...
func (s *Server) RetryCount(ctx context.Context, req *fido2.RetryCountRequest) (*fido2.RetryCountResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	count, err := device.RetryCount()
	if err != nil {
		return nil, err
	}

	return &fido2.RetryCountResponse{
		Count: int32(count),
	}, nil
}

// Assertion ...
func (s *Server) Assertion(ctx context.Context, req *fido2.AssertionRequest) (*fido2.AssertionResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	extensions, err := extensionsFromStrings(req.Extensions)
	if err != nil {
		return nil, err
	}
	uv, err := optionValueFromString(req.UV)
	if err != nil {
		return nil, err
	}
	up, err := optionValueFromString(req.UP)
	if err != nil {
		return nil, err
	}

	assertion, err := device.Assertion(req.RPID, req.ClientDataHash, req.CredentialIDs, req.PIN, &libfido2.AssertionOpts{Extensions: extensions, UV: uv, UP: up})
	if err != nil {
		return nil, err
	}

	return &fido2.AssertionResponse{
		Assertion: assertionToRPC(assertion),
	}, nil
}

// CredentialsInfo ...
func (s *Server) CredentialsInfo(ctx context.Context, req *fido2.CredentialsInfoRequest) (*fido2.CredentialsInfoResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}
	if req.PIN == "" {
		return nil, status.Error(codes.InvalidArgument, "pin required")
	}

	info, err := device.CredentialsInfo(req.PIN)
	if err != nil {
		return nil, err
	}

	return &fido2.CredentialsInfoResponse{
		Info: credentialsInfoToRPC(info),
	}, nil
}

// Credentials ...
func (s *Server) Credentials(ctx context.Context, req *fido2.CredentialsRequest) (*fido2.CredentialsResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	if req.PIN == "" {
		return nil, status.Error(codes.InvalidArgument, "pin required")
	}

	out := []*fido2.Credential{}
	if req.RPID == "" {
		rps, err := device.RelyingParties(req.PIN)
		if err != nil {
			if errors.Cause(err) == libfido2.ErrPinInvalid {
				return nil, status.Error(codes.InvalidArgument, "pin invalid")
			}
			// TODO: Bug in libfido2 or SoloKey where if there are no credentials returns ErrRXNotCBOR.
			if errors.Cause(err) != libfido2.ErrRXNotCBOR {
				return nil, err
			}
		}
		for _, rp := range rps {
			credentials, err := device.Credentials(rp.ID, req.PIN)
			if err != nil {
				return nil, err
			}
			out = append(out, credentialsToRPC(relyingPartyToRPC(rp), credentials)...)
		}
	} else {
		credentials, err := device.Credentials(req.RPID, req.PIN)
		if err != nil {
			return nil, err
		}
		rp := &fido2.RelyingParty{ID: req.RPID} // TODO: Name
		out = credentialsToRPC(rp, credentials)
	}

	return &fido2.CredentialsResponse{
		Credentials: out,
	}, nil
}

// RelyingParties ...
func (s *Server) RelyingParties(ctx context.Context, req *fido2.RelyingPartiesRequest) (*fido2.RelyingPartiesResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	if req.PIN == "" {
		return nil, status.Error(codes.InvalidArgument, "pin required")
	}

	rps, err := device.RelyingParties(req.PIN)
	if err != nil {
		return nil, err
	}

	return &fido2.RelyingPartiesResponse{
		Parties: relyingPartiesToRPC(rps),
	}, nil
}

// GenerateHMACSecret ...
func (s *Server) GenerateHMACSecret(ctx context.Context, req *fido2.GenerateHMACSecretRequest) (*fido2.GenerateHMACSecretResponse, error) {
	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	cdh := req.ClientDataHash
	if len(cdh) != 32 {
		return nil, errors.Errorf("invalid client data hash length")
	}
	if req.RP == nil {
		return nil, errors.Errorf("no rp specified")
	}
	if req.RP.ID == "" {
		return nil, errors.Errorf("empty rp id")
	}
	if req.RP.Name == "" {
		return nil, errors.Errorf("empty rp name")
	}

	if req.User == nil {
		return nil, errors.Errorf("no user specified")
	}
	if len(req.User.ID) == 0 {
		return nil, errors.Errorf("empty user id")
	}
	if req.User.Name == "" {
		return nil, errors.Errorf("empty user name")
	}

	opts := &libfido2.MakeCredentialOpts{
		Extensions: []libfido2.Extension{libfido2.HMACSecretExtension},
	}
	switch req.RK {
	case fido2.Default:
		opts.RK = libfido2.Default
	case fido2.True:
		opts.RK = libfido2.True
	case fido2.False:
		opts.RK = libfido2.False
	}

	attest, err := device.MakeCredential(
		cdh,
		rpFromRPC(req.RP),
		userFromRPC(req.User),
		libfido2.ES256, // Algorithm
		req.PIN,
		opts,
	)
	if err != nil {
		return nil, err
	}

	return &fido2.GenerateHMACSecretResponse{
		CredentialID: attest.CredentialID,
	}, nil
}

// HMACSecret ...
func (s *Server) HMACSecret(ctx context.Context, req *fido2.HMACSecretRequest) (*fido2.HMACSecretResponse, error) {
	s.Lock()
	defer s.Unlock()

	device, err := findDevice(req.Device)
	if err != nil {
		return nil, err
	}

	assertion, err := device.Assertion(
		req.RPID,
		req.ClientDataHash,
		req.CredentialIDs,
		req.PIN,
		&libfido2.AssertionOpts{
			Extensions: []libfido2.Extension{libfido2.HMACSecretExtension},
			UP:         libfido2.True,
			HMACSalt:   req.Salt,
		},
	)
	if err != nil {
		return nil, err
	}

	return &fido2.HMACSecretResponse{
		HMACSecret: assertion.HMACSecret,
	}, nil
}
