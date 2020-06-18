package mock

import (
	"github.com/google/uuid"
	"github.com/keys-pub/keys-ext/auth/fido2"
)

type device struct {
	dev         *fido2.Device
	info        *fido2.DeviceInfo
	pin         string
	credentials map[string]*credential
}

type credential struct {
	cred   *fido2.Credential
	secret []byte
}

type deviceOpts struct{}

func newDevice(path string, opts *deviceOpts) *device {
	if opts == nil {
		opts = &deviceOpts{}
	}
	dev := &fido2.Device{
		Path:         path,
		ProductID:    123,
		VendorID:     456,
		Manufacturer: "Test Co.",
		Product:      "Test Key",
	}
	info := &fido2.DeviceInfo{
		Versions:   []string{},
		Extensions: []string{"hmac-secret", "credMgmt"},
		AAGUID:     uuid.New().String(),
		Options:    []*fido2.Option{},
	}
	return &device{
		dev:         dev,
		info:        info,
		pin:         "12345",
		credentials: map[string]*credential{},
	}
}

func newCredential(cred *fido2.Credential, secret []byte) *credential {
	return &credential{
		cred:   cred,
		secret: secret,
	}
}
