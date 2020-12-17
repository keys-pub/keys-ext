package rpc

import (
	"github.com/google/uuid"
	"github.com/keys-pub/go-libfido2"
	"github.com/keys-pub/keys-ext/auth/fido2"
	"github.com/pkg/errors"
)

func devicesToRPC(ins []*libfido2.DeviceLocation) []*fido2.Device {
	outs := make([]*fido2.Device, 0, len(ins))
	for _, in := range ins {
		outs = append(outs, &fido2.Device{
			Path:         in.Path,
			ProductID:    int32(in.ProductID),
			VendorID:     int32(in.VendorID),
			Manufacturer: in.Manufacturer,
			Product:      in.Product,
		})
	}
	return outs
}

func bytesToUUIDString(b []byte) string {
	uid, err := uuid.FromBytes(b)
	if err == nil {
		return uid.String()
	}
	return ""
}

func deviceInfoToRPC(in *libfido2.DeviceInfo) *fido2.DeviceInfo {
	return &fido2.DeviceInfo{
		Versions:   in.Versions,
		Extensions: in.Extensions,
		AAGUID:     bytesToUUIDString(in.AAGUID),
		Options:    optionsToRPC(in.Options),
	}
}

func optionsToRPC(ins []libfido2.Option) []*fido2.Option {
	outs := make([]*fido2.Option, 0, len(ins))
	for _, in := range ins {
		outs = append(outs, &fido2.Option{
			Name:  in.Name,
			Value: optionValueToRPC(in.Value),
		})
	}
	return outs
}

func rpFromRPC(rp *fido2.RelyingParty) libfido2.RelyingParty {
	return libfido2.RelyingParty{
		ID:   rp.ID,
		Name: rp.Name,
	}
}

func userFromRPC(user *fido2.User) libfido2.User {
	return libfido2.User{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Icon:        user.Icon,
	}
}

func userToRPC(user libfido2.User) *fido2.User {
	return &fido2.User{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Icon:        user.Icon,
	}
}

func credentialTypeFromString(typ string) (libfido2.CredentialType, error) {
	switch typ {
	case "es256", "ES256":
		return libfido2.ES256, nil
	case "eddsa", "EDDSA":
		return libfido2.EDDSA, nil
	case "rs256", "RS256":
		return libfido2.RS256, nil
	default:
		return 0, errors.Errorf("unknown credential type %v", typ)
	}
}

func credentialTypeToString(typ libfido2.CredentialType) string {
	switch typ {
	case libfido2.ES256:
		return "es256"
	case libfido2.EDDSA:
		return "eddsa"
	case libfido2.RS256:
		return "rs256"
	default:
		return ""
	}
}

func optionValueToRPC(in libfido2.OptionValue) fido2.OptionValue {
	switch in {
	case libfido2.True:
		return fido2.True
	case libfido2.False:
		return fido2.False
	default:
		return fido2.Default
	}
}

func optionValueFromString(in string) (libfido2.OptionValue, error) {
	switch in {
	case "true":
		return libfido2.True, nil
	case "false":
		return libfido2.False, nil
	case "":
		return libfido2.Default, nil
	default:
		return "", errors.Errorf("invalid option value")
	}
}

func extensionsFromStrings(ins []string) ([]libfido2.Extension, error) {
	outs := []libfido2.Extension{}
	for _, in := range ins {
		ext, err := extensionFromString(in)
		if err != nil {
			return nil, err
		}
		outs = append(outs, ext)
	}
	return outs, nil
}

func extensionFromString(s string) (libfido2.Extension, error) {
	switch s {
	case "hmac-secret":
		return libfido2.HMACSecretExtension, nil
	case "credProtect":
		return libfido2.CredProtectExtension, nil
	default:
		return "", errors.Errorf("invalid extension %s", s)
	}
}

func attestationToRPC(in *libfido2.Attestation) *fido2.Attestation {
	return &fido2.Attestation{
		ClientDataHash: in.ClientDataHash,
		AuthData:       in.AuthData,
		CredentialID:   in.CredentialID,
		CredentialType: credentialTypeToString(in.CredentialType),
		PubKey:         in.PubKey,
		Cert:           in.Cert,
		Sig:            in.Sig,
		Format:         in.Format,
	}
}

func assertionToRPC(in *libfido2.Assertion) *fido2.Assertion {
	return &fido2.Assertion{
		// TODO: Change to AuthDataCBOR
		AuthData:   in.AuthDataCBOR,
		Sig:        in.Sig,
		HMACSecret: in.HMACSecret,
	}
}

func credentialsInfoToRPC(in *libfido2.CredentialsInfo) *fido2.CredentialsInfo {
	return &fido2.CredentialsInfo{
		RKExisting:  int32(in.RKExisting),
		RKRemaining: int32(in.RKRemaining),
	}
}

func credentialsToRPC(rp *fido2.RelyingParty, ins []*libfido2.Credential) []*fido2.Credential {
	outs := make([]*fido2.Credential, 0, len(ins))
	for _, in := range ins {
		outs = append(outs, credentialToRPC(rp, in))
	}
	return outs
}

func credentialToRPC(rp *fido2.RelyingParty, in *libfido2.Credential) *fido2.Credential {
	return &fido2.Credential{
		ID:   in.ID,
		Type: credentialTypeToString(in.Type),
		RP:   rp,
		User: userToRPC(in.User),
	}
}

func relyingPartiesToRPC(ins []*libfido2.RelyingParty) []*fido2.RelyingParty {
	outs := make([]*fido2.RelyingParty, 0, len(ins))
	for _, in := range ins {
		outs = append(outs, relyingPartyToRPC(in))
	}
	return outs
}

func relyingPartyToRPC(in *libfido2.RelyingParty) *fido2.RelyingParty {
	return &fido2.RelyingParty{
		ID:   in.ID,
		Name: in.Name,
	}
}
