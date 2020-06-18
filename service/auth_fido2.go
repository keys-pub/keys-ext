package service

import (
	"bytes"
	"context"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/auth/fido2"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

type authDevice struct {
	Device     *fido2.Device
	DeviceInfo *fido2.DeviceInfo
	Provision  *vault.Provision
}

// findDevice returns supported device.
// If infos is specified we try to find the device matching the auth credentials (aaguid).
func findDevice(ctx context.Context, auths fido2.AuthServer, provisions []*vault.Provision) (*authDevice, error) {
	if auths == nil {
		return nil, errors.Errorf("fido2 plugin not available")
	}

	devicesResp, err := auths.Devices(ctx, &fido2.DevicesRequest{})
	if err != nil {
		return nil, err
	}
	if len(devicesResp.Devices) == 0 {
		return nil, errors.Errorf("no device found")
	}

	// TODO: We return first device found, but we might want the user to choose instead.

	for _, device := range devicesResp.Devices {
		infoResp, err := auths.DeviceInfo(ctx, &fido2.DeviceInfoRequest{Device: device.Path})
		if err != nil {
			return nil, err
		}
		deviceInfo := infoResp.Info
		logger.Debugf("Checking device: %v", deviceInfo)
		if deviceInfo.HasExtension(fido2.HMACSecretExtension) {
			provision := matchAAGUID(provisions, deviceInfo.AAGUID)
			if len(provisions) == 0 || provision != nil {
				logger.Debugf("Found device: %v", device.Path)
				return &authDevice{
					Device:     device,
					DeviceInfo: deviceInfo,
					Provision:  provision,
				}, nil
			}
		}
	}
	return nil, errors.Errorf("no devices found matching our credentials")
}

func setupHMACSecret(ctx context.Context, auths fido2.AuthServer, vlt *vault.Vault, pin string, appName string) (*vault.Provision, error) {
	cdh := bytes.Repeat([]byte{0x00}, 32) // No client data
	rp := &fido2.RelyingParty{
		ID:   "keys.pub",
		Name: "keys.pub",
	}

	logger.Debugf("Auth setup hmac-secret, looking for supported devices...")
	authDevice, err := findDevice(ctx, auths, nil)
	if err != nil {
		return nil, err
	}

	userID := keys.Rand16()[:]

	// TODO: Default to using resident key?

	logger.Debugf("Generating hmac-secret...")
	resp, err := auths.GenerateHMACSecret(ctx, &fido2.GenerateHMACSecretRequest{
		Device:         authDevice.Device.Path,
		PIN:            pin,
		ClientDataHash: cdh[:],
		RP:             rp,
		User: &fido2.User{
			ID:   userID,
			Name: appName,
		},
		// RK: fido2.True,
	})
	if err != nil {
		return nil, err
	}

	noPin := false
	if pin == "" {
		noPin = true
	}

	id := encoding.MustEncode(resp.CredentialID, encoding.Base62)
	salt := keys.Rand32()
	provision := &vault.Provision{
		ID:        id,
		Type:      vault.FIDO2HMACSecretAuth,
		AAGUID:    authDevice.DeviceInfo.AAGUID,
		Salt:      salt[:],
		NoPin:     noPin,
		CreatedAt: time.Now(),
	}

	logger.Debugf("Saving provision: %v...", provision)
	if err := vlt.SaveProvision(provision); err != nil {
		return nil, err
	}

	return provision, nil
}

func hmacSecret(ctx context.Context, auths fido2.AuthServer, vlt *vault.Vault, pin string) ([]byte, *vault.Provision, error) {
	cdh := bytes.Repeat([]byte{0x00}, 32) // No client data
	rp := &fido2.RelyingParty{
		ID:   "keys.pub",
		Name: "keys.pub",
	}

	provisions, err := vlt.Provisions()
	if err != nil {
		return nil, nil, err
	}
	if len(provisions) == 0 {
		return nil, nil, errors.Errorf("no provisions found for hmac-secret")
	}

	logger.Debugf("Looking for device with a matching credential...")
	authDevice, err := findDevice(ctx, auths, provisions)
	if err != nil {
		return nil, nil, err
	}
	if authDevice.Provision == nil {
		return nil, nil, errors.Errorf("device has no provision")
	}

	credID, err := encoding.Decode(authDevice.Provision.ID, encoding.Base62)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "credential (provision) id was invalid")
	}

	logger.Debugf("Getting hmac-secret...")
	secretResp, err := auths.HMACSecret(ctx, &fido2.HMACSecretRequest{
		Device:         authDevice.Device.Path,
		PIN:            pin,
		ClientDataHash: cdh[:],
		RPID:           rp.ID,
		CredentialID:   credID,
		Salt:           authDevice.Provision.Salt,
	})
	if err != nil {
		return nil, nil, err
	}

	return secretResp.HMACSecret, authDevice.Provision, nil
}

func unlockHMACSecret(ctx context.Context, auths fido2.AuthServer, vlt *vault.Vault, pin string) error {
	secret, provision, err := hmacSecret(ctx, auths, vlt, pin)
	if err != nil {
		return err
	}
	if len(secret) != 32 {
		return errors.Errorf("invalid key length from hmac-secret")
	}
	key := keys.Bytes32(secret)

	// If we have setup hmac-secret provision but have not finished setup, we do
	// that on the first unlock. When we setup the hmac-secret, we use MakeCredential
	// which usually requires user presence (touching the device). Unlock also
	// usually requires user presence so we split up these blocking calls into
	// two requests. The first request doesn't give us the auth, so we do the
	// setup of first unlock instead of during setup.
	status, err := vlt.Status()
	if err != nil {
		return err
	}
	if status == vault.Setup {
		if err := vlt.Setup(key, provision); err != nil {
			return err
		}
	} else {
		if _, err := vlt.Unlock(key); err != nil {
			return err
		}
	}
	return nil
}

func provisionHMACSecret(ctx context.Context, auths fido2.AuthServer, vlt *vault.Vault, pin string) (*vault.Provision, error) {
	secret, provision, err := hmacSecret(ctx, auths, vlt, pin)
	if err != nil {
		return nil, err
	}
	if len(secret) != 32 {
		return nil, errors.Errorf("invalid key length for hmac secret")
	}
	key := keys.Bytes32(secret)

	if err := vlt.Provision(key, provision); err != nil {
		return nil, err
	}
	logger.Infof("Provision (hmac-secret): %s", provision.ID)
	return provision, nil
}
