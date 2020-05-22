package fido2

import "github.com/keys-pub/go-libfido2"

// Extension alias for libfido2.Extension.
type Extension = libfido2.Extension

// HMACSecretExtension alias for libfido2.HMACSecretExtension.
const HMACSecretExtension Extension = libfido2.HMACSecretExtension

// CredProtectExtension alias for libfido2.CredProtectExtension.
const CredProtectExtension Extension = libfido2.CredProtectExtension

// HasExtension returns true if device has extension.
func (d *DeviceInfo) HasExtension(ext Extension) bool {
	for _, e := range d.Extensions {
		if e == string(ext) {
			return true
		}
	}
	return false
}
