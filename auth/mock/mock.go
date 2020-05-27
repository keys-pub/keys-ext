package mock

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keysd/auth/fido2"
	"github.com/pkg/errors"
)

type server struct {
	devices []*device
}

// NewAuthServer creates an AuthServer.
func NewAuthServer() fido2.AuthServer {
	devices := []*device{
		newDevice("/dev/test", nil),
	}
	return &server{
		devices: devices,
	}
}

// Devices ...
func (s *server) Devices(ctx context.Context, req *fido2.DevicesRequest) (*fido2.DevicesResponse, error) {
	devices := make([]*fido2.Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device.dev)
	}
	return &fido2.DevicesResponse{
		Devices: devices,
	}, nil
}

func (s *server) findDevice(path string) *device {
	for _, d := range s.devices {
		if d.dev.Path == path {
			return d
		}
	}
	return nil
}

// DeviceInfo ...
func (s *server) DeviceInfo(ctx context.Context, req *fido2.DeviceInfoRequest) (*fido2.DeviceInfoResponse, error) {
	dev := s.findDevice(req.Device)
	if dev == nil {
		return &fido2.DeviceInfoResponse{}, nil
	}
	return &fido2.DeviceInfoResponse{
		Info: dev.info,
	}, nil
}

// MakeCredential ...
func (s *server) MakeCredential(ctx context.Context, req *fido2.MakeCredentialRequest) (*fido2.MakeCredentialResponse, error) {
	return nil, errors.Errorf("not implemented")
}

// SetPIN ...
func (s *server) SetPIN(ctx context.Context, req *fido2.SetPINRequest) (*fido2.SetPINResponse, error) {
	dev := s.findDevice(req.Device)
	if dev == nil {
		return nil, errors.Errorf("device not found: %s", req.Device)
	}

	if dev.pin != "" && req.OldPIN != dev.pin {
		return nil, errors.Errorf("invalid old pin")
	}

	dev.pin = req.PIN

	return &fido2.SetPINResponse{}, nil
}

// Reset ...
func (s *server) Reset(ctx context.Context, req *fido2.ResetRequest) (*fido2.ResetResponse, error) {
	return nil, errors.Errorf("not implemented")
}

// RetryCount ...
func (s *server) RetryCount(ctx context.Context, req *fido2.RetryCountRequest) (*fido2.RetryCountResponse, error) {
	return nil, errors.Errorf("not implemented")
}

// Assertion ...
func (s *server) Assertion(ctx context.Context, req *fido2.AssertionRequest) (*fido2.AssertionResponse, error) {
	return nil, errors.Errorf("not implemented")
}

// CredentialsInfo ...
func (s *server) CredentialsInfo(ctx context.Context, req *fido2.CredentialsInfoRequest) (*fido2.CredentialsInfoResponse, error) {
	return nil, errors.Errorf("not implemented")
}

// Credentials ...
func (s *server) Credentials(ctx context.Context, req *fido2.CredentialsRequest) (*fido2.CredentialsResponse, error) {
	return nil, errors.Errorf("not implemented")
}

// RelyingParties ...
func (s *server) RelyingParties(ctx context.Context, req *fido2.RelyingPartiesRequest) (*fido2.RelyingPartiesResponse, error) {
	return nil, errors.Errorf("not implemented")
}

// GenerateHMACSecret ...
func (s *server) GenerateHMACSecret(ctx context.Context, req *fido2.GenerateHMACSecretRequest) (*fido2.GenerateHMACSecretResponse, error) {
	dev := s.findDevice(req.Device)
	if dev == nil {
		return nil, errors.Errorf("device not found: %s", req.Device)
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

	cred := &fido2.Credential{
		ID:   keys.Rand32()[:],
		Type: "es256",
		RP:   req.RP,
		User: req.User,
	}

	secret := keys.Rand32()[:]
	id := encoding.MustEncode(cred.ID, encoding.Base62)
	dev.credentials[id] = newCredential(cred, secret)

	return &fido2.GenerateHMACSecretResponse{
		CredentialID: cred.ID,
	}, nil
}

// HMACSecret ...
func (s *server) HMACSecret(ctx context.Context, req *fido2.HMACSecretRequest) (*fido2.HMACSecretResponse, error) {
	dev := s.findDevice(req.Device)
	if dev == nil {
		return nil, errors.Errorf("device not found: %s", req.Device)
	}

	id := encoding.MustEncode(req.CredentialID, encoding.Base62)
	cred, ok := dev.credentials[id]
	if !ok {
		return nil, errors.Errorf("credential not found")
	}

	if len(req.Salt) != 32 {
		return nil, errors.Errorf("invalid salt")
	}

	h := hmac.New(sha256.New, bytes.Join([][]byte{cred.secret, req.Salt}, []byte{}))
	h.Write(req.ClientDataHash)
	hmacSecret := h.Sum(nil)

	return &fido2.HMACSecretResponse{
		HMACSecret: hmacSecret,
	}, nil
}
