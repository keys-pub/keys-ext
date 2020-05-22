package fido2

// Extension (FIDO2).
type Extension string

// HMACSecretExtension (should match libfido2.HMACSecretExtension).
const HMACSecretExtension Extension = "hmac-secret"

// CredProtectExtension (should match libfido2.CredProtectExtension).
const CredProtectExtension Extension = "credProtect"

// HasExtension returns true if device has extension.
func (d *DeviceInfo) HasExtension(ext Extension) bool {
	for _, e := range d.Extensions {
		if e == string(ext) {
			return true
		}
	}
	return false
}
