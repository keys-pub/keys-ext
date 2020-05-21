package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keysd/fido2"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// authCredential stores metadata and parameters used for the generated HMAC secret credential.
type authCredential struct {
	ID     []byte `msgpack:"id"`
	AAGUID string `msgpack:"aaguid"`
	Salt   []byte `msgpack:"salt"`
}

type authDevice struct {
	Device *fido2.Device
	Info   *fido2.DeviceInfo
	Cred   *authCredential
}

// findDevice returns supported device.
// If creds is specified we try to find the device matching the auth credentials (aaguid).
func (a *auth) findDevice(ctx context.Context, creds []*authCredential) (*authDevice, error) {
	devicesResp, err := a.authenticators.Devices(ctx, &fido2.DevicesRequest{})
	if err != nil {
		return nil, err
	}
	if len(devicesResp.Devices) == 0 {
		return nil, errors.Errorf("no device found")
	}

	// TODO: We return first device found, but we might want the user to choose instead.

	for _, device := range devicesResp.Devices {
		infoResp, err := a.authenticators.DeviceInfo(ctx, &fido2.DeviceInfoRequest{Device: device.Path})
		if err != nil {
			return nil, err
		}
		info := infoResp.Info
		if info.HasExtension(fido2.HMACSecretExtension) {
			cred := matchCred(creds, info.AAGUID)
			if len(creds) == 0 || cred != nil {
				return &authDevice{
					Device: device,
					Info:   info,
					Cred:   cred,
				}, nil
			}
		}
	}
	return nil, errors.Errorf("no device found matching ids")
}

func (a *auth) loadCredentials() ([]*authCredential, error) {
	st := a.keyring.Store()
	service := a.cfg.KeyringService(st.Name())
	ids, err := st.IDs(service, &keyring.IDsOpts{Prefix: "#cred-", ShowReserved: true})
	if err != nil {
		return nil, err
	}
	creds := make([]*authCredential, 0, len(ids))
	for _, id := range ids {
		b, err := st.Get(service, id)
		if err != nil {
			return nil, err
		}
		var cred authCredential
		if err := msgpack.Unmarshal(b, &cred); err != nil {
			return nil, err
		}
		creds = append(creds, &cred)
	}
	return creds, nil
}

func (a *auth) saveCredential(cred *authCredential) error {
	st := a.keyring.Store()
	service := a.cfg.KeyringService(st.Name())
	kp := fmt.Sprintf("#cred-" + keys.Rand3262())
	b, err := msgpack.Marshal(cred)
	if err != nil {
		return err
	}
	if err := st.Set(service, kp, b); err != nil {
		return err
	}
	return nil
}

func (a *auth) hmacSecret(ctx context.Context, pin string, setup bool) ([]byte, error) {
	cdh := bytes.Repeat([]byte{0x00}, 32) // No client data
	rpID := "keys.pub"

	var authDevice *authDevice
	if setup {
		authDevice, err := a.findDevice(ctx, nil)
		if err != nil {
			return nil, err
		}

		resp, err := a.authenticators.GenerateHMACSecret(ctx, &fido2.GenerateHMACSecretRequest{
			Device:         authDevice.Device.Path,
			PIN:            pin,
			ClientDataHash: cdh[:],
			RPID:           rpID,
		})
		if err != nil {
			return nil, err
		}

		salt := keys.Rand32()
		cred := &authCredential{
			AAGUID: authDevice.Info.AAGUID,
			ID:     resp.CredentialID,
			Salt:   salt[:],
		}
		if err := a.saveCredential(cred); err != nil {
			return nil, err
		}
		authDevice.Cred = cred

	} else {
		creds, err := a.loadCredentials()
		if err != nil {
			return nil, err
		}
		if len(creds) == 0 {
			return nil, errors.Errorf("no credentials found for hmac-secret")
		}
		found, err := a.findDevice(ctx, creds)
		if err != nil {
			return nil, err
		}
		authDevice = found
	}

	secretResp, err := a.authenticators.HMACSecret(ctx, &fido2.HMACSecretRequest{
		Device:         authDevice.Device.Path,
		PIN:            pin,
		ClientDataHash: cdh[:],
		RPID:           rpID,
		CredentialID:   authDevice.Cred.ID,
		Salt:           authDevice.Cred.Salt,
	})
	if err != nil {
		return nil, err
	}

	return secretResp.HMACSecret, nil
}

func (a *auth) unlockWithFIDO2HMACSecret(ctx context.Context, secret string, setup bool) (string, keyring.Auth, error) {

	key := keys.Rand32()

	auth := keyring.NewKeyAuth(key)

	if setup {
		id, err := a.keyring.Setup(auth)
		if err != nil {
			return "", nil, err
		}
		return id, auth, nil
	}

	id, err := a.keyring.Unlock(auth)
	if err != nil {
		return "", nil, err
	}
	return id, auth, nil
}

func matchCred(creds []*authCredential, aaguid string) *authCredential {
	for _, cred := range creds {
		if cred.AAGUID == aaguid {
			return cred
		}
	}
	return nil
}
