package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/vmihailenco/msgpack/v4"
)

// Encrypt to recipients saltpack.
func Encrypt(i interface{}, sender *keys.EdX25519Key, recipients ...keys.ID) ([]byte, error) {
	b, err := msgpack.Marshal(i)
	if err != nil {
		return nil, err
	}
	return saltpack.Signcrypt(b, false, sender, recipients...)
}

// Decrypt and unmarshal into value for recipient.
func Decrypt(b []byte, v interface{}, kr saltpack.Keyring) (keys.ID, error) {
	dec, pk, err := saltpack.SigncryptOpen(b, false, kr)
	if err != nil {
		return "", err
	}
	if err := msgpack.Unmarshal(dec, v); err != nil {
		return "", err
	}
	if pk != nil {
		return pk.ID(), nil
	}
	return "", nil
}
