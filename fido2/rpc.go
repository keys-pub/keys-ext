package fido2

import (
	"github.com/keys-pub/go-libfido2"
	"github.com/pkg/errors"
)

func devicesToRPC(ins []*libfido2.DeviceLocation) []*Device {
	outs := make([]*Device, 0, len(ins))
	for _, in := range ins {
		outs = append(outs, &Device{
			Path:         in.Path,
			ProductID:    int32(in.ProductID),
			VendorID:     int32(in.VendorID),
			Manufacturer: in.Manufacturer,
			Product:      in.Product,
		})
	}
	return outs
}

func deviceInfoToRPC(in *libfido2.DeviceInfo) *DeviceInfo {
	return &DeviceInfo{
		Versions:   in.Versions,
		Extensions: in.Extensions,
		AAGUID:     in.AAGUID,
		Options:    optionsToRPC(in.Options),
	}
}

func optionsToRPC(ins []libfido2.Option) []*Option {
	outs := make([]*Option, 0, len(ins))
	for _, in := range ins {
		outs = append(outs, &Option{
			Name:  in.Name,
			Value: optionValueToRPC(in.Value),
		})
	}
	return outs
}

func rpFromRPC(rp *RelyingParty) libfido2.RelyingParty {
	return libfido2.RelyingParty{
		ID:   rp.ID,
		Name: rp.Name,
	}
}

func userFromRPC(user *User) libfido2.User {
	return libfido2.User{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Icon:        user.Icon,
	}
}

func userToRPC(user libfido2.User) *User {
	return &User{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Icon:        user.Icon,
	}
}

func credTypeFromRPC(typ string) (libfido2.CredentialType, error) {
	switch typ {
	case "es256":
		return libfido2.ES256, nil
	case "eddsa":
		return libfido2.EDDSA, nil
	case "rs256":
		return libfido2.RS256, nil
	default:
		return 0, errors.Errorf("unknown credential type %v", typ)
	}
}

func credTypeToRPC(typ libfido2.CredentialType) string {
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

func optionValueToRPC(in libfido2.OptionValue) string {
	switch in {
	case libfido2.True:
		return "true"
	case libfido2.False:
		return "false"
	default:
		return ""
	}
}

func optionValueFromRPC(in string) (libfido2.OptionValue, error) {
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

func extensionsFromRPC(ins []string) ([]libfido2.Extension, error) {
	outs := []libfido2.Extension{}
	for _, in := range ins {
		ext, err := extensionFromRPC(in)
		if err != nil {
			return nil, err
		}
		outs = append(outs, ext)
	}
	return outs, nil
}

func extensionFromRPC(s string) (libfido2.Extension, error) {
	switch s {
	case "hmac-secret":
		return libfido2.HMACSecret, nil
	case "credProtect":
		return libfido2.CredProtect, nil
	default:
		return "", errors.Errorf("invalid extension %s", s)
	}
}

func attestationToRPC(in *libfido2.Attestation) *Attestation {
	return &Attestation{
		ClientDataHash: in.ClientDataHash,
		AuthData:       in.AuthData,
		CredID:         in.CredID,
		CredType:       credTypeToRPC(in.CredType),
		PubKey:         in.PubKey,
		Cert:           in.Cert,
		Sig:            in.Sig,
		Format:         in.Format,
	}
}

func assertionToRPC(in *libfido2.Assertion) *Assertion {
	return &Assertion{
		AuthData:   in.AuthData,
		Sig:        in.Sig,
		HMACSecret: in.HMACSecret,
	}
}

func credentialsInfoToRPC(in *libfido2.CredentialsInfo) *CredentialsInfo {
	return &CredentialsInfo{
		RKExisting:  int32(in.RKExisting),
		RKRemaining: int32(in.RKRemaining),
	}
}

func credentialsToRPC(ins []*libfido2.Credential) []*Credential {
	outs := make([]*Credential, 0, len(ins))
	for _, in := range ins {
		outs = append(outs, credentialToRPC(in))
	}
	return outs
}

func credentialToRPC(in *libfido2.Credential) *Credential {
	return &Credential{
		ID:   in.ID,
		Type: credTypeToRPC(in.Type),
		User: userToRPC(in.User),
	}
}

func relyingPartiesToRPC(ins []*libfido2.RelyingParty) []*RelyingParty {
	outs := make([]*RelyingParty, 0, len(ins))
	for _, in := range ins {
		outs = append(outs, relyingPartyToRPC(in))
	}
	return outs
}

func relyingPartyToRPC(in *libfido2.RelyingParty) *RelyingParty {
	return &RelyingParty{
		ID:   in.ID,
		Name: in.Name,
	}
}
