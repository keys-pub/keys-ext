package service

import (
	"bytes"
	"context"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keys-ext/auth/fido2"
	"github.com/pkg/errors"
)

type authDevice struct {
	Device     *fido2.Device
	DeviceInfo *fido2.DeviceInfo
	Info       *authInfo
}

// findDevice returns supported device.
// If infos is specified we try to find the device matching the auth credentials (aaguid).
func (a *auth) findDevice(ctx context.Context, infos []*authInfo) (*authDevice, error) {
	if a.fido2 == nil {
		return nil, errors.Errorf("fido2 plugin not available")
	}

	devicesResp, err := a.fido2.Devices(ctx, &fido2.DevicesRequest{})
	if err != nil {
		return nil, err
	}
	if len(devicesResp.Devices) == 0 {
		return nil, errors.Errorf("no device found")
	}

	// TODO: We return first device found, but we might want the user to choose instead.

	for _, device := range devicesResp.Devices {
		infoResp, err := a.fido2.DeviceInfo(ctx, &fido2.DeviceInfoRequest{Device: device.Path})
		if err != nil {
			return nil, err
		}
		deviceInfo := infoResp.Info
		if deviceInfo.HasExtension(fido2.HMACSecretExtension) {
			info := matchAAGUID(infos, deviceInfo.AAGUID)
			if len(infos) == 0 || info != nil {
				return &authDevice{
					Device:     device,
					DeviceInfo: deviceInfo,
					Info:       info,
				}, nil
			}
		}
	}
	return nil, errors.Errorf("no devices found matching our credentials")
}

func (a *auth) setupHMACSecret(ctx context.Context, pin string) (string, error) {
	cdh := bytes.Repeat([]byte{0x00}, 32) // No client data
	rp := &fido2.RelyingParty{
		ID:   "keys.pub",
		Name: "keys.pub",
	}

	logger.Debugf("Auth setup hmac-secret, looking for supported devices...")
	authDevice, err := a.findDevice(ctx, nil)
	if err != nil {
		return "", err
	}

	userID := keys.Rand16()[:]

	// TODO: Default to using resident key?

	logger.Debugf("Generating hmac-secret...")
	resp, err := a.fido2.GenerateHMACSecret(ctx, &fido2.GenerateHMACSecretRequest{
		Device:         authDevice.Device.Path,
		PIN:            pin,
		ClientDataHash: cdh[:],
		RP:             rp,
		User: &fido2.User{
			ID:   userID,
			Name: a.cfg.AppName(),
		},
		// RK: fido2.True,
	})
	if err != nil {
		return "", err
	}

	noPin := false
	if pin == "" {
		noPin = true
	}

	id := encoding.MustEncode(resp.CredentialID, encoding.Base62)
	salt := keys.Rand32()
	info := &authInfo{
		ID:        id,
		Type:      fido2HMACSecretAuth,
		AAGUID:    authDevice.DeviceInfo.AAGUID,
		Salt:      salt[:],
		NoPin:     noPin,
		CreatedAt: time.Now(),
	}
	if err := a.saveInfo(info); err != nil {
		return "", err
	}

	return id, nil
}

func (a *auth) hmacSecret(ctx context.Context, pin string) ([]byte, string, error) {
	cdh := bytes.Repeat([]byte{0x00}, 32) // No client data
	rp := &fido2.RelyingParty{
		ID:   "keys.pub",
		Name: "keys.pub",
	}

	infos, err := a.loadInfos()
	if err != nil {
		return nil, "", err
	}
	if len(infos) == 0 {
		return nil, "", errors.Errorf("no metadata found for hmac-secret")
	}

	logger.Debugf("Looking for device with a matching credential...")
	authDevice, err := a.findDevice(ctx, infos)
	if err != nil {
		return nil, "", err
	}
	if authDevice.Info == nil {
		return nil, "", errors.Errorf("device has no metadata")
	}

	credID, err := encoding.Decode(authDevice.Info.ID, encoding.Base62)
	if err != nil {
		return nil, "", errors.Wrapf(err, "credential (auth) id was invalid")
	}

	logger.Debugf("Getting hmac-secret...")
	secretResp, err := a.fido2.HMACSecret(ctx, &fido2.HMACSecretRequest{
		Device:         authDevice.Device.Path,
		PIN:            pin,
		ClientDataHash: cdh[:],
		RPID:           rp.ID,
		CredentialID:   credID,
		Salt:           authDevice.Info.Salt,
	})
	if err != nil {
		return nil, "", err
	}

	return secretResp.HMACSecret, authDevice.Info.ID, nil
}

func (a *auth) unlockHMACSecret(ctx context.Context, pin string) error {
	key, id, err := a.hmacSecret(ctx, pin)
	if err != nil {
		return err
	}
	if len(key) != 32 {
		return errors.Errorf("invalid key length from hmac-secret")
	}
	auth := keyring.NewAuth(id, keys.Bytes32(key))

	// If we have setup hmac-secret but have not setup the keyring, we do that
	// on the first unlock. When we setup the hmac-secret, we use MakeCredential
	// which usually requires user presence (touching the device). Unlock also
	// usually requires user presence so we split up these blocking calls into
	// two requests. The first request doesn't give us the auth, so we do the
	// keyring setup of first unlock instead of during setup.
	status, err := a.kr.Status()
	if err != nil {
		return err
	}
	if status == keyring.Setup {
		if _, err := a.kr.Setup(auth); err != nil {
			return err
		}
	} else {
		if _, err := a.kr.Unlock(auth); err != nil {
			return err
		}
	}
	return nil
}

func (a *auth) provisionHMACSecret(ctx context.Context, pin string) (string, error) {
	key, credID, err := a.hmacSecret(ctx, pin)
	if err != nil {
		return "", err
	}
	if len(key) != 32 {
		return "", errors.Errorf("invalid key length for hmac secret")
	}
	auth := keyring.NewAuth(credID, keys.Bytes32(key))
	id, err := a.kr.Provision(auth)
	if err != nil {
		return "", err
	}
	logger.Infof("Provision (hmac-secret) with auth id: %s", id)
	return id, nil
}
