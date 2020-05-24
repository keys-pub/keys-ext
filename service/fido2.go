package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keysd/auth/fido2"
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
	if a.auths == nil {
		return nil, errors.Errorf("fido2 plugin not available")
	}

	devicesResp, err := a.auths.Devices(ctx, &fido2.DevicesRequest{})
	if err != nil {
		return nil, err
	}
	if len(devicesResp.Devices) == 0 {
		return nil, errors.Errorf("no device found")
	}

	// TODO: We return first device found, but we might want the user to choose instead.

	for _, device := range devicesResp.Devices {
		infoResp, err := a.auths.DeviceInfo(ctx, &fido2.DeviceInfoRequest{Device: device.Path})
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
	return nil, errors.Errorf("no devices found matching our credentials")
}

func (a *auth) loadCredentials() ([]*authCredential, error) {
	st := a.keyring.Store()
	service := a.cfg.KeyringService(st.Name())
	ids, err := st.IDs(service, keyring.WithReservedPrefix("#cred-"))
	if err != nil {
		return nil, err
	}
	logger.Debugf("Looking up credentials %v", ids)
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

func (a *auth) saveCredential(cred *authCredential, credID string) error {
	st := a.keyring.Store()
	service := a.cfg.KeyringService(st.Name())
	kp := fmt.Sprintf("#cred-" + credID)
	logger.Debugf("Saving credential %s", kp)
	b, err := msgpack.Marshal(cred)
	if err != nil {
		return err
	}
	if err := st.Set(service, kp, b); err != nil {
		return err
	}
	return nil
}

func (a *auth) setupHMACSecret(ctx context.Context, pin string) error {
	cdh := bytes.Repeat([]byte{0x00}, 32) // No client data
	rp := &fido2.RelyingParty{
		ID:   "keys.pub",
		Name: a.cfg.AppName(),
	}

	logger.Debugf("Auth setup hmac-secret, looking for supported devices...")
	authDevice, err := a.findDevice(ctx, nil)
	if err != nil {
		return err
	}

	logger.Debugf("Generating hmac-secret...")
	resp, err := a.auths.GenerateHMACSecret(ctx, &fido2.GenerateHMACSecretRequest{
		Device:         authDevice.Device.Path,
		PIN:            pin,
		ClientDataHash: cdh[:],
		RP:             rp,
		User: &fido2.User{
			ID:   []byte("-"),
			Name: "-",
		},
	})
	if err != nil {
		return err
	}

	salt := keys.Rand32()
	cred := &authCredential{
		AAGUID: authDevice.Info.AAGUID,
		ID:     resp.CredentialID,
		Salt:   salt[:],
	}
	credID := encoding.MustEncode(resp.CredentialID, encoding.Base62)
	if err := a.saveCredential(cred, credID); err != nil {
		return err
	}

	return nil
}

func (a *auth) hmacSecret(ctx context.Context, pin string) ([]byte, error) {
	cdh := bytes.Repeat([]byte{0x00}, 32) // No client data
	rp := &fido2.RelyingParty{
		ID:   "keys.pub",
		Name: a.cfg.AppName(),
	}

	creds, err := a.loadCredentials()
	if err != nil {
		return nil, err
	}
	if len(creds) == 0 {
		return nil, errors.Errorf("no credentials found for hmac-secret")
	}

	logger.Debugf("Looking for device with a matching credential...")
	authDevice, err := a.findDevice(ctx, creds)
	if err != nil {
		return nil, err
	}
	if authDevice.Cred == nil {
		return nil, errors.Errorf("device has no credentials")
	}

	logger.Debugf("Getting hmac-secret...")
	secretResp, err := a.auths.HMACSecret(ctx, &fido2.HMACSecretRequest{
		Device:         authDevice.Device.Path,
		PIN:            pin,
		ClientDataHash: cdh[:],
		RPID:           rp.ID,
		CredentialID:   authDevice.Cred.ID,
		Salt:           authDevice.Cred.Salt,
	})
	if err != nil {
		return nil, err
	}

	return secretResp.HMACSecret, nil
}

func (a *auth) unlockHMACSecret(ctx context.Context, pin string) (keyring.Auth, error) {
	key, err := a.hmacSecret(ctx, pin)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.Errorf("invalid key length from hmac-secret")
	}
	auth := keyring.NewKeyAuth(keys.Bytes32(key))

	// If we have setup hmac-secret but have not setup the keyring, we do that
	// on the first unlock. When we setup the hmac-secret, we use MakeCredential
	// which usually requires user presence (touching the device). On unlock
	// also usually requires user presence so we split up these blocking calls
	// into two requests. The first request doesn't give us the auth, so we do
	// the keyring setup of first unlock instead of during setup.
	isSetup, err := a.keyring.IsSetup()
	if err != nil {
		return nil, err
	}
	if !isSetup {
		if _, err := a.keyring.Setup(auth); err != nil {
			return nil, err
		}
	} else {
		if _, err := a.keyring.Unlock(auth); err != nil {
			return nil, err
		}
	}
	return auth, nil
}

func matchCred(creds []*authCredential, aaguid string) *authCredential {
	for _, cred := range creds {
		if cred.AAGUID == aaguid {
			return cred
		}
	}
	return nil
}
